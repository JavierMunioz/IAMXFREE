# IAMXFREE

IAMXFREE is a terminal UI (TUI) for administering applications deployed on a
VPS — an alternative to bare SSH commands, in the spirit of `lazydocker`,
`k9s` and `btop`. Everything runs inside the terminal; there is no graphical
component.

## Status

Early scaffolding. The interactive dashboard, application registration flow
and infrastructure managers (Nginx, Apache, process supervision, etc.) do not
exist yet — this repository currently only proves the base plumbing:
Cobra-driven CLI → Bubble Tea TUI shell.

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

| Package                  | Responsibility                                                          |
| ------------------------ | ------------------------------------------------------------------------ |
| `cmd/iamxfree`            | Entrypoint binary. Stays thin; delegates to `internal/cli`.             |
| `internal/cli`            | Command/flag parsing (Cobra). No business logic.                        |
| `internal/tui`            | Presentation layer (Bubble Tea/Lipgloss). Renders state, emits intents. |
| `internal/core`           | Orchestrates services/managers/repositories for each use case.          |
| `internal/managers`       | Concrete resource managers: processes, Nginx, Apache, env files, etc.   |
| `internal/services`       | Business logic coordinating managers + repositories.                    |
| `internal/models`         | Domain entities (Application, Deployment, ...).                         |
| `internal/repositories`   | Persistence of IAMXFREE's own state.                                    |
| `internal/adapters`       | Integration with external systems (systemd, git, ssh, ...).             |
| `internal/infrastructure` | Cross-cutting concerns: logging, process execution, filesystem.         |
| `internal/config`         | IAMXFREE's own configuration loading/validation.                        |
| `internal/plugins`        | Reserved extension point for future technologies.                       |

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
