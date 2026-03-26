package skills

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// Skill is a piece of contextual knowledge an agent can request on demand.
// Examples: a guide for writing async services, coding standards, API docs.
type Skill interface {
	// Name returns a short unique identifier (used to request the skill).
	Name() string
	// Description returns a one-line summary shown to the agent in its prompt.
	Description() string
	// Load returns the full text content of the skill.
	Load(ctx context.Context) (string, error)
}

// FileSkill loads skill content from a file on disk.
type FileSkill struct {
	name        string
	description string
	path        string
}

func NewFileSkill(name, description, path string) *FileSkill {
	return &FileSkill{name: name, description: description, path: path}
}

func (s *FileSkill) Name() string        { return s.name }
func (s *FileSkill) Description() string { return s.description }
func (s *FileSkill) Load(_ context.Context) (string, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return "", fmt.Errorf("load skill %q: %w", s.name, err)
	}
	return string(data), nil
}

// StaticSkill holds skill content inline (useful for tests or configuration).
type StaticSkill struct {
	name        string
	description string
	content     string
}

func NewStaticSkill(name, description, content string) *StaticSkill {
	return &StaticSkill{name: name, description: description, content: content}
}

func (s *StaticSkill) Name() string                            { return s.name }
func (s *StaticSkill) Description() string                     { return s.description }
func (s *StaticSkill) Load(_ context.Context) (string, error)  { return s.content, nil }

// Registry holds a named set of skills and provides lookup by name.
type Registry struct {
	skills map[string]Skill
	order  []string // preserve insertion order for listing
}

func NewRegistry(ss ...Skill) *Registry {
	r := &Registry{skills: make(map[string]Skill)}
	for _, s := range ss {
		r.skills[s.Name()] = s
		r.order = append(r.order, s.Name())
	}
	return r
}

// Get retrieves a skill by name.
func (r *Registry) Get(name string) (Skill, bool) {
	s, ok := r.skills[name]
	return s, ok
}

// List returns a formatted bullet list of available skills (name + description)
// to be injected into the agent's system prompt.
func (r *Registry) List() string {
	if len(r.order) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("\n\n---\n## Available Skills\n\n")
	sb.WriteString("You can request any of the following reference documents using the `get_skill` tool:\n\n")
	for _, name := range r.order {
		s := r.skills[name]
		fmt.Fprintf(&sb, "- **%s**: %s\n", s.Name(), s.Description())
	}
	return sb.String()
}

// Empty reports whether the registry has no skills.
func (r *Registry) Empty() bool { return len(r.order) == 0 }
