package factory

import (
	"log/slog"

	"github.com/oskov/project-pipe/internal/agent"
	"github.com/oskov/project-pipe/internal/llm"
	"github.com/oskov/project-pipe/internal/service"
	"github.com/oskov/project-pipe/internal/store"
	"github.com/oskov/project-pipe/internal/tools/toolsets"
)

// ManagerFactory returns a managerFactory function suitable for TaskService.
// It is exported so main.go can reuse it when wiring the worker dispatcher.
func ManagerFactory(s store.Store, llmClient llm.Client) func(string, *slog.Logger) service.ManagerAgent {
	ticketSvc := service.NewTicketService(s.Tickets())
	memorySvc := service.NewMemoryService(s.AgentMemory())

	return func(projectID string, taskLogger *slog.Logger) service.ManagerAgent {
		// Workspace services — nil until GitHub integration provides a workDir.
		var fsSvc service.FilesystemService
		var goSvc service.GoToolchainService
		var parseSvc service.GoParseService

		golangDeveloper := agent.NewGolangDeveloper(llmClient, s.AgentRuns(),
			agent.WithTools(toolsets.GolangDeveloperTools(memorySvc, projectID, fsSvc, goSvc, parseSvc)...),
			agent.WithLogger(taskLogger),
		)
		devManager := agent.NewDevManager(llmClient, s.AgentRuns(),
			agent.WithTools(toolsets.DevManagerTools(memorySvc, projectID, parseSvc, golangDeveloper)...),
			agent.WithLogger(taskLogger),
		)
		architect := agent.NewArchitect(llmClient, s.AgentRuns(),
			agent.WithTools(toolsets.ArchitectTools(memorySvc, projectID, fsSvc, parseSvc)...),
			agent.WithLogger(taskLogger),
		)
		return agent.NewManager(llmClient, s.AgentRuns(),
			agent.WithTools(toolsets.ManagerTools(ticketSvc, memorySvc, projectID, architect, devManager)...),
			agent.WithLogger(taskLogger),
		)
	}
}
