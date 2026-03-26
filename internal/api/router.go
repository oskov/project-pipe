package api

import (
"context"
"log/slog"
"net/http"

"github.com/go-chi/chi/v5"
chimiddleware "github.com/go-chi/chi/v5/middleware"
"github.com/oskov/project-pipe/internal/agent"
apimiddleware "github.com/oskov/project-pipe/internal/api/middleware"
"github.com/oskov/project-pipe/internal/config"
"github.com/oskov/project-pipe/internal/llm"
"github.com/oskov/project-pipe/internal/service"
"github.com/oskov/project-pipe/internal/store"
"github.com/oskov/project-pipe/internal/tools/toolsets"
)

// ManagerFactory returns a managerFactory function suitable for TaskService.
// It resolves the project's local_path from the DB and wires up all
// workspace-scoped services for the cloned repository.
func ManagerFactory(s store.Store, llmClient llm.Client, projectSvc service.ProjectService, githubToken string) func(string, string, *slog.Logger) service.ManagerAgent {
ticketSvc := service.NewTicketService(s.Tickets())
memorySvc := service.NewMemoryService(s.AgentMemory())
prSvc := service.NewPullRequestService(s.PullRequests(), s.Projects(), githubToken)

return func(taskID, projectID string, taskLogger *slog.Logger) service.ManagerAgent {
var workDir string
if p, err := projectSvc.GetByID(context.Background(), projectID); err == nil {
workDir = p.LocalPath
}

var fsSvc service.FilesystemService
var goSvc service.GoToolchainService
var parseSvc service.GoParseService
var gitSvc service.GitService

if workDir != "" {
fsSvc = service.NewFilesystemService(workDir)
goSvc = service.NewGoToolchainService(workDir)
parseSvc = service.NewGoParseService(workDir)
gitSvc = service.NewGitService(workDir, githubToken)
}

golangDeveloper := agent.NewGolangDeveloper(llmClient, s.AgentRuns(),
agent.WithTools(toolsets.GolangDeveloperTools(memorySvc, taskID, projectID, fsSvc, goSvc, parseSvc, gitSvc, prSvc)...),
agent.WithLogger(taskLogger),
)
devManager := agent.NewDevManager(llmClient, s.AgentRuns(),
agent.WithTools(toolsets.DevManagerTools(memorySvc, taskID, projectID, parseSvc, prSvc, golangDeveloper)...),
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
func NewRouter(s store.Store, llmClient llm.Client, gitCfg config.GitConfig, taskSvc service.TaskService, logger *slog.Logger) http.Handler {
r := chi.NewRouter()

r.Use(chimiddleware.RequestID)
r.Use(chimiddleware.Recoverer)
r.Use(apimiddleware.StructuredLogger(logger))

projectSvc := service.NewProjectService(s.Projects(), gitCfg.ProjectsDir, gitCfg.GithubToken)
prSvc := service.NewPullRequestService(s.PullRequests(), s.Projects(), gitCfg.GithubToken)

th := &taskHandler{tasks: taskSvc}
prh := &taskPRsHandler{prs: prSvc}
ph := &projectHandler{projects: projectSvc}

r.Route("/api/v1", func(r chi.Router) {
r.Post("/tasks", th.createTask)
r.Get("/tasks/{id}", th.getTask)
r.Get("/tasks/{id}/prs", prh.listTaskPRs)

r.Post("/projects", ph.createProject)
r.Get("/projects", ph.listProjects)
r.Get("/projects/{id}", ph.getProject)
})

return r
}
