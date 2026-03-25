# Go Developer Workflow

## Iterative Development Cycle

Always follow this loop when implementing changes:

1. **Read first** — use `list_files` and `read_file` to understand existing code before writing anything.
2. **Write** — use `write_file` to implement the change.
3. **Build** — run `go_command` with subcommand `build` and args `["./..."]`. Fix any errors before proceeding.
4. **Vet** — run `go_command` with subcommand `vet` and args `["./..."]`.
5. **Test** — run `go_command` with subcommand `test` and args `["./..."]`. Fix failures.
6. **Repeat** until build and tests pass cleanly.

## Key Rules

- Never leave the codebase in a broken state (failing build or tests).
- Always run `go build ./...` after every file change.
- Use `go help <topic>` if unsure about a Go feature or flag.
- Prefer small, focused changes — one concern per file change.
- Follow existing code style (formatting, naming, error handling patterns).

## Error Handling

- Wrap errors with `fmt.Errorf("context: %w", err)` — never discard errors.
- Return errors up the call stack; only log at the top level.

## Reporting

When done, provide a summary:
- Files changed (created/modified/deleted)
- What each change does
- Build and test status
