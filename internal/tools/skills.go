package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oskov/project-pipe/internal/skills"
)

// GetSkill is a tool that fetches the full content of a named skill document
// from the registry on demand.
type GetSkill struct {
	registry *skills.Registry
}

func NewGetSkill(registry *skills.Registry) *GetSkill {
	return &GetSkill{registry: registry}
}

func (t *GetSkill) Name() string        { return "get_skill" }
func (t *GetSkill) Description() string { return "Fetch the full content of a reference document (skill) by name. Use this to read guides, standards, or other reference materials listed in the Available Skills section." }
func (t *GetSkill) Parameters() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"name": {"type": "string", "description": "The exact skill name as listed in Available Skills"}
		},
		"required": ["name"]
	}`)
}

func (t *GetSkill) Execute(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	skill, ok := t.registry.Get(args.Name)
	if !ok {
		return fmt.Sprintf("skill %q not found. Available skills are listed in your system prompt.", args.Name), nil
	}

	content, err := skill.Load(ctx)
	if err != nil {
		return "", err
	}
	return content, nil
}
