# IAMXFREE

IAMXFREE is a terminal UI (TUI) for administering applications deployed on a
VPS — an alternative to bare SSH commands, in the spirit of `lazydocker`,
`k9s` and `btop`. Everything runs inside the terminal; there is no graphical
component.

## Status

The main dashboard, the application-registration wizard, and a full
application detail/administration screen are implemented. Node.js
applications managed with npm can now actually be started and stopped by
IAMXFREE itself from that detail screen (see "Execution engine" below) —
every other technology (Python, Go, PHP, Docker, systemd, PM2, ...) still
has no Execution Strategy yet, so their applications show no live
Strategy/Health/session data, only what's stored.

Registering an application analyzes the project path first: the wizard asks
for a path, inspects it, shows what it found, and pre-fills name, type,
framework, runtime, package manager and install/build/start commands from
that analysis — every field stays editable, and nothing is ever guessed
silently (see "Application Setup Service" below).

From the dashboard:
- `a` register a new application · `enter` open the selected card's detail screen
- arrows / `tab` / `shift+tab` move the selection · `r` refresh the list · `q` quit

From an application's detail screen (see "Application Detail Screen" below):
- `s` start · `x` stop · `f5` refresh health/session
- `r` restart · `l` logs · `e` edit config (all three: not implemented yet)
- `b` / `esc` back to the dashboard · `q` quit

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
| `internal/tui/dashboard`             | The main screen: card grid of registered applications, top bar — the entry point. |
| `internal/tui/detail`                | Per-application administration screen: diagnose, start, stop, refresh.          |
| `internal/tui/wizard`                | Generic, feature-agnostic multi-step form engine used by every TUI wizard.       |
| `internal/tui/wizards/application`   | Composes the wizard engine into the concrete "create application" flow.         |
| `internal/validation`                | Reusable, composable input validators (Required, Port, Domain, URL, ...).        |
| `internal/core`                      | Orchestrates services/managers/repositories for each use case.                  |
| `internal/execution`                 | Technology-agnostic contract for installing/building/running an application.    |
| `internal/inspection`                | Reads a directory and detects what kind of project lives there. Read-only.      |
| `internal/planner`                   | Turns an inspection.Result into a proposed IAMXFREE configuration.              |
| `internal/runtimehost`               | The only package allowed to touch os/exec — running commands, checking PATH.   |
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
grid. The dashboard itself now holds no per-application actions at all —
pressing Enter emits `OpenDetailMsg{AppID}` and whatever hosts it (today,
`internal/tui`'s `RootModel`) decides what that means; the dashboard has no
notion of "detail view" beyond that message, the same pattern `a` already
used for `OpenWizardMsg`. This is deliberate: the dashboard is a list and an
entry point, nothing else.

When an application resolves to a registered `execution.Strategy` (Node+npm
today), its card grows two extra lines — `Strategy: <name>` and a
`Health:`/icon line using the same reserved status-color convention as the
run-status icon — fetched once per `Reload()` via
`ApplicationService.CheckExecutionHealth` and simply omitted for
applications with no resolvable strategy yet.

### Application Detail Screen

`internal/tui/detail` is where a single registered application is actually
administered — diagnosed, started, and stopped. It is its own top-level
screen (a sibling of the dashboard and the wizard, not a mode nested inside
either), opened from the dashboard's `OpenDetailMsg` and returned from via
its own `BackMsg`. Like every other screen, it depends only on
`services.ApplicationService` and the new `services.ExecutionService` —
never on `internal/execution`, `internal/runtimehost`, or a repository
directly (verified with
`grep -rn "internal/execution\|internal/runtimehost\|internal/repositories" internal/tui/detail/*.go`,
which matches nothing outside a doc comment).

Four panels:
- **Top** — general info: name, type, framework, runtime, resolved
  `Strategy`, the application's stored `Status`, live `Health`, port, domain.
- **Left ("Technical")** — where it lives and how it's built: path,
  repository, package manager, install/build/start commands, registration
  date.
- **Right ("Execution")** — the live session IAMXFREE is tracking, if any:
  running/stopped, PID, start time, working directory, runtime. When
  nothing has been started yet (or IAMXFREE was restarted, losing its
  in-memory tracking — see "Runtime Host"), this says so explicitly rather
  than showing blank fields.
- **Bottom** — the action bar, listing every keybinding the screen responds
  to, including the ones that don't do anything real yet (marked with `*`):
  `s` start, `x` stop, `r`\* restart, `l`\* logs, `e`\* edit config, `f5`
  refresh, `b`/`esc` back, `q` quit. No dead keys — every one of them
  produces a visible status message.

The session isn't persisted anywhere: pressing `s` calls
`ExecutionService.Start`, and the resulting `services.RunSession` (a plain
projection of `execution.Session` — PID, start time, command, args, working
dir, status, runtime) is held only in the screen's own in-memory state for
as long as it stays open. Pressing `x` calls `ExecutionService.Stop` and
clears it. `s` refuses to start a second session over a tracked one (it
would silently leak the first process's PID); `x` refuses when nothing is
tracked. `f5` is the *only* refresh — there is no automatic/periodic
polling yet — and it re-runs both the health check and, if a session is
tracked, `ExecutionService.RefreshSession` (backed by the new
`Strategy.Status` method, see "Execution engine") to notice if the process
died on its own since the screen last checked.

### Runtime Host

`internal/runtimehost` is the only package allowed to talk to the operating
system — no other package imports `os/exec` (verified with
`grep -rl '"os/exec"' --include="*.go" . | grep -v internal/runtimehost`,
which returns nothing outside it). Execution strategies depend on it; it
never depends on them.

- **`Host`** — the interface: `LookPath` (is a tool on PATH — "not found" is
  a normal result, not an error), `Version` (run a tool's version flag and
  report it structurally), `Run`/`RunCaptured` (synchronous), `ReadFile`,
  `WorkingDir`, `FileExists`/`DirExists`, and `StartProcess`/
  `IsProcessRunning`/`StopProcess` for a long-running background process
  (no supervision/auto-restart yet — that's a later iteration).
- **`LinuxHost`** — the real implementation, backed by `os/exec` and `os`.
  `StartProcess` tracks the resulting `*os.Process` and reaps it in a
  background goroutine once it exits, so a long-running child (a web
  server, say) never becomes a zombie even though nothing calls `Wait()`
  explicitly; `IsProcessRunning`/`StopProcess` fall back to a raw
  PID-based OS check when a PID isn't one this Host instance started
  itself.
- **`runtimehosttest.FakeHost`** — a builder-style test double
  (`WithLookPath`/`WithVersion`/`WithRunResult`/`WithReadFile`/
  `WithStartProcess`/`WithRunningPID`/`WithStopError`/...) in its own
  importable subpackage (not `_test.go`), so execution strategies' tests
  depend on deterministic, configured responses instead of whatever happens
  to be installed on the machine running them.

Structured models instead of bare errors: `CommandResult` (exit code,
captured stdout/stderr, duration), `ToolAvailability`/`ToolInfo` (found vs.
not found is data, not a failure), and `ExecutionError` (implements `error`
and `Unwrap`, carrying command/args/exit code/stderr so a caller can inspect
a failure without re-parsing a message string).

### Execution engine

`internal/execution` defines *how* an application is installed, built,
started, stopped, restarted and updated. Three pieces:

- **`Strategy`** — the contract one technology (Node+npm, Python+uv, Docker
  Compose, systemd, ...) implements: `CanHandle(app)` (pure, no I/O),
  `Metadata()`, `HealthCheck`/`Readiness` (diagnostics), `Start`/`Stop`
  (session-aware execution), `Status` (re-checks whether a previously
  returned `Session` is still alive — a point-in-time query, never
  mutating or persisting anything), and `Install`/`Build`/`Restart`/`Update`
  (still `execution.ErrNotImplemented` for every strategy, Node included —
  out of scope so far).
- **`Registry`** — where strategies register themselves (`Register(strategy)`).
  Adding a new technology never means editing existing code, only
  constructing its `Strategy` and registering it.
- **`Resolver`** — given an `Application`, asks each registered strategy (in
  registration order) whether it can handle it, and returns the first match.
  No if/else chain to extend as technologies are added.

Shared models every `Strategy` uses the same way (not per-technology
concepts): `Status` (`starting`/`running`/`stopping`/`stopped`/`failed`,
room to grow), `Session` (never just a PID — start time, command, args,
working dir, status, runtime), `HealthCheck`/`HealthCheckItem` (a fixed,
technology-agnostic vocabulary of check names —
`runtime_installed`/`package_manager_installed`/`manifest_exists`/
`path_valid`/`scripts_available`/`commands_configured`/
`directory_accessible` — each a structured pass/fail, never free text; what
a name means in practice is technology-specific and goes in `Detail`), and
`Readiness` (`DeriveReadiness(health)` is a single shared policy: a failed
dependency check becomes a `MissingDependencies` entry, `scripts_available`
failing is downgraded to a `Warnings` entry since it isn't always fatal,
and anything else failed is a `BlockingErrors` entry — every one of those
except warnings flips `Ready` to false).

**Node.js (npm) is the reference `Strategy` implementation** — every future
technology follows its shape. `HealthCheck` verifies node/npm are on PATH,
the configured path is set and exists, `package.json` exists and declares at
least one script (read via `Host.ReadFile` + a minimal local JSON struct —
strategies never import `internal/inspection`, so this doesn't reuse the
detector's parsing code; detection and execution stay decoupled on purpose),
and a start command is configured. `Start` calls `Readiness` first and
refuses if it isn't ready, splits the configured start command on whitespace
(not a full shell parser — quoted arguments with embedded spaces aren't
supported yet) and runs it via `Host.StartProcess`, returning a `Session`.
`Stop` calls `Host.StopProcess(session.PID)`. `Status` calls
`Host.IsProcessRunning(session.PID)` and returns the same `Session` with
just its `Status` field refreshed. `CanHandle` matches only
`Runtime == node && Config.PackageManager == "npm"` — a Node app with no
package manager configured, or a different one (pnpm, yarn, bun), is left
for another `Strategy` to claim later, never assumed to be npm.

Two services sit between all of this and `internal/tui`, split by concern
rather than crammed into one:

- **`ApplicationService`** — CRUD on the `Application` entity, plus
  `ResolveExecutionStrategy` (returns a strategy's `Metadata` without
  invoking anything) and `CheckExecutionHealth` (actually runs
  `HealthCheck`, summarized into `services.ExecutionHealth{StrategyName,
  Healthy}`).
- **`ExecutionService`** (new) — controlling a running application:
  `Start`/`Stop`/`RefreshSession`, all working in terms of
  `services.RunSession` (a plain projection of `execution.Session`) so
  callers never need to import `internal/execution` to hold onto one
  between calls. Both services resolve the same application's `Strategy`
  independently through their own `*execution.Resolver`, since neither
  depends on the other.

`internal/cli/root.go` registers `execution.NewNodeStrategy` (backed by
`runtimehost.NewLinuxHost()`) into the registry at startup, and constructs
both `ApplicationService` and `ExecutionService` from the same resolver.

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
