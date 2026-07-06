# IAMXFREE

IAMXFREE is a terminal UI (TUI) for administering applications deployed on a
VPS — an alternative to bare SSH commands, in the spirit of `lazydocker`,
`k9s` and `btop`. Everything runs inside the terminal; there is no graphical
component.

## Status

Early scaffolding. The application-registration wizard is implemented and
persists to the JSON store; the card/panel dashboard and infrastructure
managers (Nginx, Apache, process supervision, etc.) do not exist yet.

From the running TUI: press `a` to register a new application, `q` to quit.

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
| `internal/tui/wizard`                | Generic, feature-agnostic multi-step form engine used by every TUI wizard.       |
| `internal/tui/wizards/application`   | Composes the wizard engine into the concrete "create application" flow.         |
| `internal/validation`                | Reusable, composable input validators (Required, Port, Domain, URL, ...).        |
| `internal/core`                      | Orchestrates services/managers/repositories for each use case.                  |
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

The wizard never persists anything itself. Whatever hosts it (today,
`internal/tui`'s `RootModel`) converts its `Result` into a domain draft, then
calls the relevant service (e.g. `ApplicationService.Register`), which
validates, checks for conflicts, and persists through the repository layer.

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
