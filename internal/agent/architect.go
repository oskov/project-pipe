package agent

import (
	"embed"
	"log/slog"

	"github.com/oskov/project-pipe/internal/llm"
	"github.com/oskov/project-pipe/internal/skills"
	"github.com/oskov/project-pipe/internal/store"
)

//go:embed architect_skills/*
var architectSkillsFS embed.FS

const architectSystemPrompt = `You are a software architect agent in a development pipeline.
You receive a structured task from the manager agent, explore the codebase, and produce a detailed
architectural change plan for the developer agent. Follow the Architecture Planning Guide.`

// NewArchitect creates an Architect agent with embedded default skills.
func NewArchitect(llmClient llm.Client, agentRuns store.AgentRunRepository, opts ...Option) *Agent {
	registry, err := skills.RegistryFromFS(architectSkillsFS, "architect_skills")
	if err != nil {
		slog.Warn("failed to load architect default skills", "error", err)
		registry = skills.NewRegistry()
	}
	return New(store.AgentTypeArchitect, architectSystemPrompt, llmClient, agentRuns,
		append([]Option{WithSkills(registry)}, opts...)...)
}
