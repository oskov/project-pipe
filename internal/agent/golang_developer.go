package agent

import (
	"embed"
	"log/slog"

	"github.com/oskov/project-pipe/internal/llm"
	"github.com/oskov/project-pipe/internal/skills"
	"github.com/oskov/project-pipe/internal/store"
)

//go:embed golang_developer_skills/*
var golangDeveloperSkillsFS embed.FS

const golangDeveloperSystemPrompt = `You are a senior Go developer agent. You receive an architectural plan and implement it.

Your workflow MUST be:
1. Read existing code with list_files and read_file to understand the codebase.
2. Implement changes with write_file.
3. Run go_command build to verify the code compiles. Fix all errors.
4. Run go_command vet to catch issues.
5. Run go_command test to verify correctness. Fix all failures.
6. Repeat steps 3–5 until everything passes cleanly.

Never stop until the build and all tests pass. Report what you changed and the final build/test status.`

// NewGolangDeveloper creates a Go developer agent with embedded default skills.
func NewGolangDeveloper(llmClient llm.Client, agentRuns store.AgentRunRepository, opts ...Option) *Agent {
	registry, err := skills.RegistryFromFS(golangDeveloperSkillsFS, "golang_developer_skills")
	if err != nil {
		slog.Warn("failed to load golang developer default skills", "error", err)
		registry = skills.NewRegistry()
	}
	return New(store.AgentTypeGolangDeveloper, golangDeveloperSystemPrompt, llmClient, agentRuns,
		append([]Option{WithSkills(registry)}, opts...)...)
}
