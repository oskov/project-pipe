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

// NewRouter builds and returns the chi router with all routes registered.
func NewRouter(s store.Store, llmClient llm.Client, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Recoverer)
	r.Use(apimiddleware.StructuredLogger(logger))

	managerFactory := func(projectID string) *agent.Agent {
		golangDeveloper := agent.NewGolangDeveloper(llmClient, s.AgentRuns(),
			agent.WithTools(toolsets.GolangDeveloperTools("", s.AgentMemory(), projectID)...),
			agent.WithLogger(logger),
		)
		devManager := agent.NewDevManager(llmClient, s.AgentRuns(),
			agent.WithTools(toolsets.DevManagerTools(s.AgentMemory(), projectID, "", golangDeveloper)...),
			agent.WithLogger(logger),
		)
		architect := agent.NewArchitect(llmClient, s.AgentRuns(),
			agent.WithTools(toolsets.ArchitectTools("", s.AgentMemory(), projectID)...),
			agent.WithLogger(logger),
		)
		return agent.NewManager(llmClient, s.AgentRuns(),
			agent.WithTools(toolsets.ManagerTools(
				s.Tickets(), s.AgentMemory(), projectID,
				architect, devManager,
			)...),
			agent.WithLogger(logger),
		)
	}

	th := &taskHandler{tasks: service.NewTaskService(s.Tasks(), s.Projects(), managerFactory)}
	ph := &projectHandler{projects: service.NewProjectService(s.Projects())}

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/tasks", th.createTask)

		r.Post("/projects", ph.createProject)
		r.Get("/projects", ph.listProjects)
		r.Get("/projects/{id}", ph.getProject)
	})

	return r
}
