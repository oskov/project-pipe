package agent

import (
	"embed"
	"log/slog"

	"github.com/oskov/project-pipe/internal/llm"
	"github.com/oskov/project-pipe/internal/skills"
	"github.com/oskov/project-pipe/internal/store"
)

//go:embed manager_skills/*
var managerSkillsFS embed.FS

const managerSystemPrompt = `You are a project manager agent in a software development pipeline.

When you receive a request, you MUST:
1. Call list_tickets to check existing open tickets in the project.
2. Decide: does this request match an existing ticket, or is it new work?
3. If it matches an existing ticket, call get_ticket to read its details and incorporate the new request into your analysis.
4. If it is new work, call create_ticket with a clear title and full description (requirements, acceptance criteria, constraints).
5. Return a structured task summary including the ticket ID.

Always follow the Task Analysis Guide for structuring your output.`

// NewManager creates a Manager agent with embedded default skills.
func NewManager(llmClient llm.Client, agentRuns store.AgentRunRepository, opts ...Option) *Agent {
	registry, err := skills.RegistryFromFS(managerSkillsFS, "manager_skills")
	if err != nil {
		slog.Warn("failed to load manager default skills", "error", err)
		registry = skills.NewRegistry()
	}
	return New(store.AgentTypeManager, managerSystemPrompt, llmClient, agentRuns,
		append([]Option{WithSkills(registry)}, opts...)...)
}

