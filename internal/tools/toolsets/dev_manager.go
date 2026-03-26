package toolsets

import (
"github.com/oskov/project-pipe/internal/service"
"github.com/oskov/project-pipe/internal/store"
"github.com/oskov/project-pipe/internal/tools"
)

// DevManagerTools returns the full set of tools for the developer manager agent.
// parseSvc and prSvc are workspace-scoped; pass nil when workDir is not yet available.
func DevManagerTools(
memorySvc service.MemoryService,
taskID string,
projectID string,
parseSvc service.GoParseService,
prSvc service.PullRequestService,
golangDeveloperRunner tools.AgentRunner,
) []tools.Tool {
tt := []tools.Tool{
tools.NewMemorySave(memorySvc, projectID, store.AgentTypeDevManager),
tools.NewMemoryGet(memorySvc, projectID, store.AgentTypeDevManager),
tools.NewMemoryList(memorySvc, projectID, store.AgentTypeDevManager),
tools.NewRunAgent(
"golang_developer",
"Delegate Go implementation work to the Go developer agent. Provide the architectural plan, working directory, and expected deliverables. The agent will implement changes, run go build/test/vet, and return a summary.",
projectID,
golangDeveloperRunner,
),
}
if parseSvc != nil {
tt = append(tt,
tools.NewGoDefinitions(parseSvc),
tools.NewGoReadDefinition(parseSvc),
)
}
if prSvc != nil {
tt = append(tt, tools.NewCreatePR(taskID, projectID, prSvc))
}
return tt
}
