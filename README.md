# project-pipe

A platform for automating software development using a hierarchy of LLM agents.

## Idea

You send a task in plain text. A chain of agents handles it end to end:

```
Manager → Architect → Dev Manager → Go Developer (→ more stacks soon)
```

- **Manager** — receives the task, creates or links a ticket, then delegates
- **Architect** — explores the codebase and produces a change plan
- **Dev Manager** — reads the plan and routes work to the right developer
- **Go Developer** — implements the plan, runs `go build`/`go test`/`go vet`, fixes errors, repeats until green

## Status

Early stage. The agent pipeline and core infrastructure are in place; GitHub integration (cloning repos, creating branches and PRs) is next.

## Stack

- **Go** — no CGO, single binary
- **SQLite** — via `modernc.org/sqlite` (pure Go)
- **LLM** — abstracted; currently backed by [langchaingo](https://github.com/tmc/langchaingo) (OpenAI, Anthropic, Ollama)
- **chi v5** — HTTP router

## Quick Start

```bash
cp config.yaml config.local.yaml
# set your LLM API key in config.local.yaml

go run ./cmd/server
```

Create a project, then send a task:

```bash
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{"name": "my-service", "github_repo": "owner/my-service"}'

curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"project_id": "<id>", "prompt": "Add health check endpoint"}'
```

## Configuration

```yaml
server:
  port: 8080
  log_level: info

db:
  dsn: ./project-pipe.db

llm:
  provider: openai   # openai | anthropic | ollama
  model: gpt-4o
  api_key: ${OPENAI_API_KEY}
```

## Project Structure

```
internal/
  agent/       — agent constructors and per-agent embedded skills
  api/         — HTTP handlers and router
  llm/         — LLM client abstraction
  service/     — business logic
  skills/      — lazy-loaded reference documents for agents
  store/       — domain types and repository interfaces
    sqlite/    — SQLite implementation
  tools/       — agent tools (filesystem, memory, tickets, go commands, delegation)
    toolsets/  — per-agent tool bundles
migrations/    — SQL migration files
```
