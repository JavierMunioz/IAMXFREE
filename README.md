# IAMXFREE

IAMXFREE is a terminal UI (TUI) for administering applications deployed on a
VPS — an alternative to bare SSH commands, in the spirit of `lazydocker`,
`k9s` and `btop`. Everything runs inside the terminal; there is no graphical
component.

## Status

The main dashboard and the application-registration wizard are implemented;
infrastructure managers (Nginx, Apache, process supervision, systemd, Docker,
etc.) do not exist yet — application status shown today comes from the
stored record, not a live process.

Registering an application analyzes the project path first: the wizard asks
for a path, inspects it, shows what it found, and pre-fills name, type,
framework, runtime, package manager and install/build/start commands from
that analysis — every field stays editable, and nothing is ever guessed
silently (see "Application Setup Service" below).

From the running TUI:
- `a` register a new application · `enter` open the selected card's detail view
- arrows / `tab` / `shift+tab` move the selection · `r` refresh the list
- `e` / `d` are reserved for edit/delete (not implemented yet) · `q` quit

## Stack

- **Language:** Go
- **TUI:** [Bubble Tea](https://github.com/charmbracelet/bubbletea) +
  [Lipgloss](https://github.com/charmbracelet/lipgloss)
- **CLI:** [Cobra](https://github.com/spf13/cobra)

Go was chosen because it matches the reference tools this project takes
inspiration from (`k9s`, `lazydocker`), compiles to a single dependency-free
binary that is trivial to drop onto any Linux VPS, and has first-class
support for concurrency, process management and SSH — all central to what
IAMXFREE needs to do.

## Architecture

The codebase is split by responsibility so that new technologies (a new
process supervisor, a new reverse proxy, a new plugin) can be added without
reshaping existing code:

| Package                             | Responsibility                                                                   |
| ------------------------------------ | --------------------------------------------------------------------------------- |
| `cmd/iamxfree`                       | Entrypoint binary. Stays thin; delegates to `internal/cli`.                      |
| `internal/cli`                       | Command/flag parsing (Cobra). Wires repositories/services and starts the TUI.    |
| `internal/tui`                       | Presentation layer (Bubble Tea/Lipgloss). Renders state, emits intents.          |
| `internal/tui/dashboard`             | The main screen: card grid of registered applications, top bar, detail view.    |
| `internal/tui/wizard`                | Generic, feature-agnostic multi-step form engine used by every TUI wizard.       |
| `internal/tui/wizards/application`   | Composes the wizard engine into the concrete "create application" flow.         |
| `internal/validation`                | Reusable, composable input validators (Required, Port, Domain, URL, ...).        |
| `internal/core`                      | Orchestrates services/managers/repositories for each use case.                  |
| `internal/execution`                 | Technology-agnostic contract for installing/building/running an application.    |
| `internal/inspection`                | Reads a directory and detects what kind of project lives there. Read-only.      |
| `internal/planner`                   | Turns an inspection.Result into a proposed IAMXFREE configuration.              |
| `internal/managers`                  | Concrete resource managers: processes, Nginx, Apache, env files, etc.           |
| `internal/services`                  | Business logic coordinating managers + repositories (e.g. ApplicationService).  |
| `internal/models`                    | Domain entities (Application, ApplicationDraft, ...).                           |
| `internal/repositories`              | Persistence contracts for IAMXFREE's own state.                                |
| `internal/repositories/jsonstore`    | JSON-file-backed implementation of those contracts (one file per entity).       |
| `internal/adapters`                  | Integration with external systems (systemd, git, ssh, ...).                     |
| `internal/infrastructure`            | Cross-cutting concerns: logging, process execution, filesystem.                 |
| `internal/config`                    | IAMXFREE's own configuration loading/validation.                                |
| `internal/plugins`                   | Reserved extension point for future technologies.                               |

### Wizards

`internal/tui/wizard` is a generic engine: it sequences a list of `Step`
values and hands back a `Result` once the last one is confirmed. It has no
notion of "application", "Nginx" or anything else domain-specific — it only
knows `Step.Focus/Update/View/Validate/Value/Modal`. Reusable step kinds
(`TextStep`, `ChoiceStep`, `SummaryStep`) live in that same package. Each
concrete wizard (creating an application today; Nginx, SSL, systemd units,
etc. later) gets its own subpackage under `internal/tui/wizards/` that
composes those generic steps and maps the resulting `Result` into a
feature-specific draft type — the engine itself never needs to change to add
a new wizard or a new step to an existing one.

`TextStep`/`ChoiceStep` also support `WithPrefill(func() string)`: a value
source re-evaluated on every focus but only applied while the field still
holds whatever it last auto-filled — so a value the user actually edited is
never overwritten, while an untouched field stays in sync if the upstream
source changes (e.g. the user goes back and picks a different project path).
The engine itself still doesn't know *why* a value is being suggested; that
knowledge lives entirely in the composing wizard package.

The wizard never persists anything itself. Whatever hosts it (today,
`internal/tui`'s `RootModel`) converts its `Result` into a domain draft, then
calls the relevant service (e.g. `ApplicationService.Register`), which
validates, checks for conflicts, and persists through the repository layer.

The create-application wizard's own flow: **path first**, then an
`AnalysisStep` that calls `ApplicationSetupService.Inspect` (see below) and
shows what it found, then Name/Type/Framework/Runtime/Package
manager/Install/Build/Start — each pre-filled from that analysis via
`WithPrefill` — then Port/Domain/Repository (never pre-filled) and a final
summary. `AnalysisStep` caches by path, so going back and forth between
steps without changing the path never re-runs inspection; changing the path
and returning does.

### Dashboard

`internal/tui/dashboard` is the main screen: a top status bar (hostname, OS,
application count, a live clock — server uptime is a placeholder until a
future iteration reads it for real), a grid of cards (one per registered
application, showing name/type/framework/runtime/status/port/domain), and a
bottom keybinding bar. It is an independent component: it depends only on
`services.ApplicationService`, never on a repository directly, and holds no
business logic — loading, sorting and formatting data is all it does.

Selection is never ambiguous: the active card gets both a distinct border
shape (thick vs. rounded) and an accent color, so it reads correctly even on
terminals with no color support. Status colors follow the project's
data-viz palette convention — a small fixed scale (good/warning/critical)
that always pairs a color with an icon and a label, never color alone.
Zero registered applications shows a friendly empty state instead of a blank
grid. Pressing Enter opens a read-only detail view of the selected
application; `e` (edit) and `d` (delete) are wired up but just report "not
implemented yet" for now — the keys exist so the experience is already
defined, even before the behavior behind them is.

### Execution engine

`internal/execution` defines *how* an application would be installed, built,
started, stopped, restarted and updated — without committing to any
technology or running anything. Three pieces:

- **`Strategy`** — the contract one technology (Node+npm, Python+uv, Docker
  Compose, systemd, ...) implements: `CanHandle(app)` (a pure yes/no,
  no I/O), `Metadata()` (name, supported runtimes/frameworks, requirements,
  `Capabilities`), and the six lifecycle methods. No concrete strategy exists
  yet — every lifecycle method a future strategy adds should return
  `execution.ErrNotImplemented` until it has real logic.
- **`Registry`** — where strategies register themselves (`Register(strategy)`).
  Adding a new technology never means editing existing code, only
  constructing its `Strategy` and registering it.
- **`Resolver`** — given an `Application`, asks each registered strategy (in
  registration order) whether it can handle it, and returns the first match.
  No if/else chain to extend as technologies are added.

`ApplicationService.ResolveExecutionStrategy` is the only integration point
today: it looks up an application and asks the resolver which strategy would
manage it, returning that strategy's `Metadata` — never invoking any
lifecycle method. `internal/cli/root.go` currently wires up an empty
`Registry`, so this always reports `execution.ErrNoStrategyFound` until the
first concrete strategy is registered in a later iteration.

### Project Inspector

`internal/inspection` reads a directory and reports what it finds — it never
runs a command, never installs anything, and never writes to disk. Same
shape as the execution engine, applied to detection instead of running:

- **`Detector`** — one per technology (`node_detector.go`, `python_detector.go`,
  `go_detector.go`, `php_detector.go`, `rust_detector.go`, `java_detector.go`,
  `docker_detector.go`). Each only reads files the `Inspector` already knows
  exist (via `DetectionInput.Has`/`ReadFile`) and returns a `Detection` — the
  project's runtime, package manager, framework (only when it can genuinely
  be inferred), matched marker files, suggested commands, and a `Confidence`.
  Whatever a `Detector` cannot determine goes into `Notes` instead of being
  guessed.
- **`Registry`** — where detectors register themselves; adding a technology
  never means changing existing code.
- **`Inspector`** — lists a directory once and asks every registered
  `Detector` what it recognizes. A `Result` can hold more than one
  `Detection` (a Node project with a `Dockerfile` is both), and `Primary()`
  picks the highest-confidence one without discarding the rest.
  `NewDefaultRegistry()` wires up all seven built-in detectors.

Integrated via `services.ApplicationSetupService` (see below) — the wizard
itself never imports this package directly.

### Deployment Planner

`internal/planner` turns an `inspection.Result` into a `DeploymentPlan` — a
proposed configuration (suggested name, application type, framework,
runtime, package manager, suggested port, install/build/start commands,
matched files, detected dependencies, a confidence level, warnings and
notes). The Inspector detects; the Planner interprets; neither one runs a
command, writes a file, or persists anything. It depends only on
`internal/models` and `internal/inspection` — never on the TUI, the wizard,
the dashboard, or the execution engine.

Same three-piece shape again:

- **`Planner`** — one per technology (`node_planner.go` is the only one
  implemented so far; `python_planner.go`, `go_planner.go`, `docker_planner.go`
  etc. will follow the same pattern). `CanPlan(detection)` decides whether it
  applies; `Plan(detection, result)` builds the proposal. It never fails —
  anything it cannot confidently determine is left blank with a matching
  entry in `Warnings` explaining why, never guessed.
- **`Registry`** — where planners register themselves; adding a technology
  never means changing existing code. `NewDefaultRegistry()` currently
  registers just the Node planner.
- **`DeploymentPlanner`** — takes a full `inspection.Result`, picks its
  `Primary()` detection to drive the proposal, and delegates to whichever
  registered `Planner.CanPlan` matches. It always returns a `DeploymentPlan`,
  even when nothing was detected or no planner matches yet — the plan's
  `Warnings` say why instead of the call failing.

The Node planner reuses the Node detector's own install/build/start
suggestions (the detector already knows never to invent a script package.json
doesn't have) and adds proposal-level interpretation the detector doesn't do:
classifying `models.ApplicationType` (only for frameworks it's confident
about — React/Vue/Angular/Svelte → frontend, Express/NestJS → backend; hybrid
meta-frameworks like Next.js/Nuxt/Astro are deliberately left unclassified),
inferring a port from the `"start"` script's `--port`/`-p`/`PORT=` flag when
present, noting a sibling Docker/Compose detection, and downgrading its own
`Confidence` by one step whenever it had to raise a warning.

Integrated via `services.ApplicationSetupService`, same as the Inspector —
the wizard never imports `internal/planner` directly either.

### Application Setup Service

`services.ApplicationSetupService` is the orchestration layer the Wizard
requirement asked for: the only thing standing between the create-application
wizard and `internal/inspection`/`internal/planner`. `Inspect(ctx, path)`
runs the Inspector, feeds its `Result` to the `DeploymentPlanner`, and
projects the resulting `DeploymentPlan` into `ApplicationSetupProposal` — a
type that lives in `internal/services` itself, built only from
`models`/plain values, so `internal/tui` and its wizard packages never need
to import inspection or planner types at all. The Wizard's `AnalysisStep`
calls this service exactly once per distinct path and shows the user
precisely what was detected, what was inferred, and what's left blank with
a matching warning — never silent guessing.

`internal/` is used deliberately: nothing here is meant to be imported by
other Go modules, which keeps the project free to change its internals
without worrying about external consumers.

## Development

```bash
go build -o ./bin/iamxfree ./cmd/iamxfree
./bin/iamxfree
```

Requires a real terminal (TTY) to run — it will fail fast if launched from a
non-interactive shell, which is expected.

## Working conventions

This project is built iteratively: each change is designed before it is
implemented, kept small, and committed on its own using
[Conventional Commits](https://www.conventionalcommits.org/).
