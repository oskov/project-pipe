# Design: Task State Machine with Agent Checkpointing

## Problem

The current execution model is fully synchronous and stateless:

```
Worker claims task
  → TaskService.Execute()
      → Manager.react() [LLM loop]
          → run_architect tool → Architect.react() [nested LLM loop]
          → run_dev_manager tool → DevManager.react() [nested LLM loop]
              → GolangDeveloper.react() [nested LLM loop]
```

All of this happens in a single call stack. If the process is stopped while
the Manager is mid-loop (e.g., between calling the Architect and calling the
DevManager), the task is lost — on restart it will be picked up from scratch.
There is no way to resume from where execution stopped.

## Requirements

1. A task must track which agent is currently processing it.
2. On graceful shutdown, the current agent finishes its work but no new
   agent transition is initiated.
3. On restart, the task continues from the last completed agent step.

## Key Challenge

Sub-agent calls happen as **tool calls inside the Manager's ReAct loop**.
The Manager decides dynamically — via LLM reasoning — when and whether to
call the Architect or DevManager. The LLM conversation history (messages list)
lives entirely in memory.

To checkpoint between agent transitions we must:
- Persist the Manager's in-progress message history after each sub-agent tool
  call returns.
- On resume, reload that message history and continue the ReAct loop from
  where it left off.

## Proposed Design

### 1. Agent Steps Table

A new table `task_steps` tracks the execution of each named sub-agent call
within a task. It is the source of truth for task progress.

```sql
CREATE TABLE task_steps (
    id           TEXT PRIMARY KEY,
    task_id      TEXT NOT NULL REFERENCES tasks(id),
    agent_type   TEXT NOT NULL,         -- manager, architect, dev_manager, golang_developer
    step_index   INTEGER NOT NULL,      -- 0-based order within this task
    status       TEXT NOT NULL DEFAULT 'pending',  -- pending | running | completed | failed
    input        TEXT NOT NULL DEFAULT '',         -- prompt sent to the agent
    output       TEXT NOT NULL DEFAULT '',         -- result returned by the agent
    created_at   DATETIME NOT NULL,
    started_at   DATETIME,
    finished_at  DATETIME,

    UNIQUE (task_id, step_index)
);
```

### 2. Manager Conversation Checkpoint Table

To resume the Manager's ReAct loop between sub-agent calls, its message
history must be persisted.

```sql
CREATE TABLE task_manager_state (
    task_id        TEXT PRIMARY KEY REFERENCES tasks(id),
    messages_json  TEXT NOT NULL,    -- JSON array of llm.Message
    updated_at     DATETIME NOT NULL
);
```

### 3. Revised Task Statuses

```
created       -- queued, waiting for a worker
running       -- a worker has claimed it; an agent is actively executing
paused        -- shutdown was requested; current agent finished; resumable
completed     -- pipeline finished successfully
failed        -- unrecoverable error
```

`paused` is the new status written when a graceful shutdown is requested
and the in-progress sub-agent call returns cleanly.

### 4. Shutdown Signal Propagation

The worker needs a way to tell the Manager "finish your current sub-agent
call, then stop". This is done with a **shutdown channel** injected into the
execution context:

```go
// In context key package:
type shutdownKey struct{}

func ContextWithShutdown(ctx context.Context, ch <-chan struct{}) context.Context
func ShutdownChanFromContext(ctx context.Context) <-chan struct{}
```

The worker selects on both `ctx.Done()` (hard cancel) and the shutdown
channel (soft stop). When shutdown is requested:

1. Worker closes the shutdown channel.
2. The Manager's ReAct loop checks the channel **after every tool call
   returns** (i.e., between agent transitions).
3. If the channel is closed:
   - Save current message history to `task_manager_state`.
   - Set task status to `paused`.
   - Return a sentinel `ErrPaused`.
4. Worker receives `ErrPaused` — not an error, just a pause signal.
5. Worker exits without marking the task failed.

### 5. Resume Logic

When a worker claims a `paused` task:

1. Load `task_manager_state.messages_json` for the task.
2. Instantiate Manager with those pre-loaded messages (skip the initial
   system + user message injection, instead restore from DB).
3. Continue the ReAct loop from the next iteration.

Sub-agent results are **already recorded in `task_steps`**. The
`RunAgent` tool checks `task_steps` before running:

```go
// Inside RunAgent.Execute():
if step, _ := stepRepo.GetByTaskAndAgent(ctx, taskID, agentType); step.Status == "completed" {
    return step.Output, nil  // replay from cache — no LLM call
}
```

This means even if messages are replayed from a checkpoint, sub-agent calls
that already completed will return their cached output instantly.

### 6. Execution Flow (New)

```
Worker claims task (status: created or paused)
  │
  ├─ paused? → load messages from task_manager_state
  │
  ▼
Manager.resume(ctx, messages) or Manager.run(ctx, prompt)
  │
  [ReAct loop]
  │   LLM call
  │   tool call: run_architect
  │   │  RunAgent checks task_steps → not done → Architect.react()
  │   │  Architect completes → saved to task_steps
  │   ◄── tool result returned to Manager
  │
  │   [CHECK SHUTDOWN CHANNEL] ← key checkpoint
  │   if closed → save messages to DB, set status=paused, return ErrPaused
  │
  │   LLM call
  │   tool call: run_dev_manager
  │   │  ...
  │   ◄── tool result
  │
  │   [CHECK SHUTDOWN CHANNEL]
  │
  │   LLM final answer → task status=completed
```

### 7. Changes Required

#### New packages / files
- `internal/store/task_step.go` — `TaskStep` model + `TaskStepRepository` interface
- `internal/store/sqlite/task_steps.go` — SQLite implementation
- `internal/store/task_manager_state.go` — `TaskManagerStateRepository` interface
- `internal/store/sqlite/task_manager_state.go`
- `migrations/000004_task_steps.up.sql`

#### Modified files
- `internal/agent/agent.go`
  - `react()` accepts a pre-seeded messages slice (for resume)
  - After each tool-call batch, check shutdown channel from context
  - On shutdown signal, return `ErrPaused` (not a real error)
  - `Run()` → `RunOrResume(ctx, taskID, projectID, prompt, savedMessages)`
- `internal/tools/delegate.go` (`RunAgent`)
  - Before running: check `task_steps` for a completed step with this agent name
  - If found: return cached output (replay)
  - After running: save output to `task_steps`
- `internal/service/tasks.go`
  - `Execute()` saves/loads manager state from DB
  - Returns `ErrPaused` as a non-failure outcome → sets task to `paused`
- `internal/worker/dispatcher.go`
  - On soft-shutdown signal: close shutdown channel in context, do NOT cancel
    the execution context
  - `ErrPaused` from Execute → log "task paused", not an error
- `internal/store/task.go`
  - Add `TaskStatusPaused` status
  - `ClaimNext()` also picks up `paused` tasks

#### Store interface additions
```go
type TaskStepRepository interface {
    Upsert(ctx context.Context, step *TaskStep) error
    GetCompleted(ctx context.Context, taskID, agentType string) (*TaskStep, error)
    ListByTask(ctx context.Context, taskID string) ([]*TaskStep, error)
}

type TaskManagerStateRepository interface {
    Save(ctx context.Context, taskID string, messagesJSON string) error
    Load(ctx context.Context, taskID string) (string, error) // "" if not found
    Delete(ctx context.Context, taskID string) error
}
```

## What Does NOT Change

- Worker pool size, polling, per-task log files — unchanged.
- Sub-agent internal ReAct loops — they run to completion and are not
  interrupted. Only the Manager loop is checkpointed.
- The LLM interface, toolsets, skills system — unchanged.
- `GET /api/v1/tasks/{id}` returns the new `paused` status so clients can
  observe it.

## Open Questions

1. **Message history size**: Manager messages grow with each iteration. For
   long-running tasks, `messages_json` could become large (tens of KB). This
   is acceptable given the existing 32 KB tool output limit; the overall
   message history should stay in that range.

2. **Replay fidelity**: When resumed, the Manager sees the same tool results
   as before (from `task_steps`). The LLM may produce slightly different
   reasoning given a restored context, but tool outputs will be identical.

3. **Multiple workers + same paused task**: `ClaimNext` already prevents two
   workers from running the same task concurrently. A paused task is treated
   identically to a created task for claiming purposes.

4. **Nested sub-agents (DevManager → GolangDeveloper)**: The same
   checkpoint mechanism can be applied recursively to DevManager's loop if
   needed. That is a separate, additive change.

## Implementation Phases

**Phase 1** (this PR scope): Task steps table + RunAgent caching.
- Adds `task_steps`, records every sub-agent call and its output.
- `RunAgent` replays from cache on resume.
- No checkpoint of Manager messages yet — task restarts from the beginning
  but sub-agent calls that already ran are instant replays.

**Phase 2**: Manager message checkpointing.
- Add `task_manager_state` table.
- Shutdown channel in context.
- `paused` status.
- Resume from saved messages.

This phased approach lets us validate the step-recording infrastructure
(Phase 1) before adding the more complex message serialization (Phase 2).
