package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/oskov/project-pipe/internal/service"
	"github.com/oskov/project-pipe/internal/store"
)

// MemorySave saves a key-value pair to the agent's project-scoped memory.
type MemorySave struct {
	svc       service.MemoryService
	projectID string
	agentType store.AgentType
}

func NewMemorySave(svc service.MemoryService, projectID string, agentType store.AgentType) *MemorySave {
	return &MemorySave{svc: svc, projectID: projectID, agentType: agentType}
}

func (t *MemorySave) Name() string        { return "memory_save" }
func (t *MemorySave) Description() string { return "Persist a key-value entry to long-term project memory. Use this to remember important information across tasks (e.g. current architecture, decisions, module structure)." }
func (t *MemorySave) Parameters() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"key":   {"type": "string", "description": "Unique identifier for this memory entry (e.g. 'architecture', 'api-design')"},
			"value": {"type": "string", "description": "Content to store (markdown or plain text)"}
		},
		"required": ["key", "value"]
	}`)
}

func (t *MemorySave) Execute(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}
	if err := t.svc.Set(ctx, t.projectID, t.agentType, args.Key, args.Value); err != nil {
		return "", fmt.Errorf("save memory: %w", err)
	}
	return fmt.Sprintf("saved memory entry %q", args.Key), nil
}

// MemoryGet retrieves a single memory entry by key.
type MemoryGet struct {
	svc       service.MemoryService
	projectID string
	agentType store.AgentType
}

func NewMemoryGet(svc service.MemoryService, projectID string, agentType store.AgentType) *MemoryGet {
	return &MemoryGet{svc: svc, projectID: projectID, agentType: agentType}
}

func (t *MemoryGet) Name() string        { return "memory_get" }
func (t *MemoryGet) Description() string { return "Retrieve a previously saved memory entry by key." }
func (t *MemoryGet) Parameters() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"key": {"type": "string", "description": "The memory key to retrieve"}
		},
		"required": ["key"]
	}`)
}

func (t *MemoryGet) Execute(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Key string `json:"key"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}
	value, found, err := t.svc.Get(ctx, t.projectID, t.agentType, args.Key)
	if err != nil {
		return "", fmt.Errorf("get memory: %w", err)
	}
	if !found {
		return fmt.Sprintf("no memory entry found for key %q", args.Key), nil
	}
	return value, nil
}

// List lists all memory keys (with a short preview) for this agent/project.
type MemoryList struct {
	svc       service.MemoryService
	projectID string
	agentType store.AgentType
}

func NewMemoryList(svc service.MemoryService, projectID string, agentType store.AgentType) *MemoryList {
	return &MemoryList{svc: svc, projectID: projectID, agentType: agentType}
}

func (t *MemoryList) Name() string        { return "memory_list" }
func (t *MemoryList) Description() string { return "List all memory keys saved for this project, with a short preview of each value." }
func (t *MemoryList) Parameters() json.RawMessage {
	return json.RawMessage(`{"type": "object", "properties": {}}`)
}

func (t *MemoryList) Execute(ctx context.Context, _ string) (string, error) {
	entries, err := t.svc.List(ctx, t.projectID, t.agentType)
	if err != nil {
		return "", fmt.Errorf("list memory: %w", err)
	}
	if len(entries) == 0 {
		return "no memory entries found for this project", nil
	}
	var sb strings.Builder
	for _, e := range entries {
		preview := e.Value
		if len(preview) > 80 {
			preview = preview[:80] + "…"
		}
		sb.WriteString(fmt.Sprintf("- %s: %s\n", e.Key, preview))
	}
	return sb.String(), nil
}
