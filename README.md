# Day Planner

Day Planner is a locally-run, terminal-based daily planning tool for a solo developer. It enforces a structured morning ritual — reviewing tasks, setting daily goals, and scoring focus — along with intraday nudges and an end-of-day review, building the habit of intentional, deep work rather than reactive channel-checking.

## Architecture

The project follows a clean layered architecture with strict separation of concerns. No lower layer may import a higher one.

| Layer | Package | Responsibility |
|---|---|---|
| Core / Domain | `internal/core` | Domain types (`DayPlan`, `Task`), `PlannerService` interface, `Store` interface. Zero imports from other internal layers. |
| Integrations | `internal/integrations` | Adapter per external task source (Taskwarrior, Jira). May import `core`; must not import `view`. |
| View | `internal/view` | `Presenter` interface and TUI implementation. May import `core`; must not import `integrations`. |
| Entry Point | `cmd/day-planner` | Wires all layers together. The only package permitted to import from every layer. |

## How to build

```
go build ./cmd/day-planner
```

## How to run

```
./day-planner
```

## How to run tests

```
go test ./...
```
