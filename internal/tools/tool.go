package tools

import (
	"context"
	"encoding/json"

	"github.com/oskov/project-pipe/internal/llm"
)

// Tool is a capability the agent can invoke (function/tool calling).
type Tool interface {
	Name() string
	Description() string
	// Parameters returns a JSON Schema object describing the tool's arguments.
	Parameters() json.RawMessage
	// Execute runs the tool with the given JSON-encoded arguments.
	Execute(ctx context.Context, argsJSON string) (string, error)
}

// ToDefinition converts a Tool to an llm.ToolDefinition.
func ToDefinition(t Tool) llm.ToolDefinition {
	return llm.ToolDefinition{
		Name:        t.Name(),
		Description: t.Description(),
		Parameters:  t.Parameters(),
	}
}

// ToDefinitions converts a slice of Tools to llm.ToolDefinitions.
func ToDefinitions(tt []Tool) []llm.ToolDefinition {
	defs := make([]llm.ToolDefinition, len(tt))
	for i, t := range tt {
		defs[i] = ToDefinition(t)
	}
	return defs
}
