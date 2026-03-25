package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

// AgentRunner is implemented by any agent that can be invoked as a sub-agent.
// agent.Agent satisfies this interface via its Run method.
type AgentRunner interface {
	Run(ctx context.Context, taskID, projectID, prompt string) (string, error)
}

// RunAgent is a tool that delegates work to a named sub-agent and returns
// its response. The calling agent receives the full sub-agent output as the
// tool result and can use it in subsequent reasoning steps.
type RunAgent struct {
	// name is used as the tool name suffix: "run_<name>"
	name      string
	desc      string
	projectID string
	runner    AgentRunner
}

// NewRunAgent creates a delegation tool for the given sub-agent.
// name should be a short identifier, e.g. "architect" or "developer".
// desc should explain what the sub-agent does and when to call it.
func NewRunAgent(name, desc, projectID string, runner AgentRunner) *RunAgent {
	return &RunAgent{name: name, desc: desc, projectID: projectID, runner: runner}
}

func (t *RunAgent) Name() string        { return "run_" + t.name }
func (t *RunAgent) Description() string { return t.desc }
func (t *RunAgent) Parameters() json.RawMessage {
	return json.RawMessage(fmt.Sprintf(`{
		"type": "object",
		"properties": {
			"prompt": {
				"type": "string",
				"description": "Full instructions or context for the %s agent. Be specific — the agent has no memory of the current conversation."
			}
		},
		"required": ["prompt"]
	}`, t.name))
}

func (t *RunAgent) Execute(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Prompt string `json:"prompt"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}
	if args.Prompt == "" {
		return "", fmt.Errorf("prompt is required")
	}

	taskID := TaskIDFromContext(ctx)
	result, err := t.runner.Run(ctx, taskID, t.projectID, args.Prompt)
	if err != nil {
		return fmt.Sprintf("agent %s failed: %s", t.name, err.Error()), nil
	}
	return result, nil
}
