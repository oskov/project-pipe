package agent

import (
	"embed"
	"log/slog"

	"github.com/oskov/project-pipe/internal/llm"
	"github.com/oskov/project-pipe/internal/skills"
	"github.com/oskov/project-pipe/internal/store"
)

//go:embed dev_manager_skills/*
var devManagerSkillsFS embed.FS

const devManagerSystemPrompt = `You are a developer manager agent. You receive an architectural plan and delegate its implementation to the appropriate specialist developer agents.

Your responsibilities:
1. Analyse the architectural plan to determine which technology stacks are involved.
2. Select the correct developer agent for each part of the plan (e.g. run_golang_developer for Go code).
3. Pass each developer the relevant portion of the plan with full context.
4. Collect their results and compile a unified implementation summary.

Consult the Routing Guide skill to decide which developer to use for each stack.`

// NewDevManager creates a developer manager agent with embedded default skills.
func NewDevManager(llmClient llm.Client, agentRuns store.AgentRunRepository, opts ...Option) *Agent {
	registry, err := skills.RegistryFromFS(devManagerSkillsFS, "dev_manager_skills")
	if err != nil {
		slog.Warn("failed to load dev manager default skills", "error", err)
		registry = skills.NewRegistry()
	}
	return New(store.AgentTypeDevManager, devManagerSystemPrompt, llmClient, agentRuns,
		append([]Option{WithSkills(registry)}, opts...)...)
}
