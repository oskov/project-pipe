package api

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	apimiddleware "github.com/oskov/project-pipe/internal/api/middleware"
	"github.com/oskov/project-pipe/internal/llm"
	"github.com/oskov/project-pipe/internal/service"
	"github.com/oskov/project-pipe/internal/store"
)

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
