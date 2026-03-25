package api

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/oskov/project-pipe/internal/agent"
	apimiddleware "github.com/oskov/project-pipe/internal/api/middleware"
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

// NewRouter builds and returns the chi router with all routes registered.
// taskSvc is shared with the worker dispatcher so both use the same TaskService.
func NewRouter(s store.Store, llmClient llm.Client, taskSvc service.TaskService, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Recoverer)
	r.Use(apimiddleware.StructuredLogger(logger))

	th := &taskHandler{tasks: taskSvc}
	ph := &projectHandler{projects: service.NewProjectService(s.Projects())}

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/tasks", th.createTask)
		r.Get("/tasks/{id}", th.getTask)

		r.Post("/projects", ph.createProject)
		r.Get("/projects", ph.listProjects)
		r.Get("/projects/{id}", ph.getProject)
	})

	return r
}
