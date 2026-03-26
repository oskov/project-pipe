package toolsets

import (
"github.com/oskov/project-pipe/internal/service"
"github.com/oskov/project-pipe/internal/store"
"github.com/oskov/project-pipe/internal/tools"
)

// GolangDeveloperTools returns the full set of tools for the Go developer agent.
// Workspace-scoped services (fsSvc, goSvc, parseSvc, gitSvc, prSvc) may be nil
// when workDir is not yet available; those tool groups are simply omitted.
func GolangDeveloperTools(
memorySvc service.MemoryService,
taskID string,
projectID string,
fsSvc service.FilesystemService,
goSvc service.GoToolchainService,
parseSvc service.GoParseService,
gitSvc service.GitService,
prSvc service.PullRequestService,
) []tools.Tool {
tt := []tools.Tool{
tools.NewMemorySave(memorySvc, projectID, store.AgentTypeGolangDeveloper),
tools.NewMemoryGet(memorySvc, projectID, store.AgentTypeGolangDeveloper),
tools.NewMemoryList(memorySvc, projectID, store.AgentTypeGolangDeveloper),
}
if fsSvc != nil {
tt = append(tt,
tools.NewReadFile(fsSvc),
tools.NewReadFileRange(fsSvc),
tools.NewWriteFile(fsSvc),
tools.NewListFiles(fsSvc),
tools.NewSearchCode(fsSvc),
)
}
if goSvc != nil {
tt = append(tt, tools.NewGoCommand(goSvc))
}
if parseSvc != nil {
tt = append(tt,
tools.NewGoDefinitions(parseSvc),
tools.NewGoReadDefinition(parseSvc),
)
}
if gitSvc != nil {
tt = append(tt,
tools.NewGitStatus(gitSvc),
tools.NewGitDiff(gitSvc),
tools.NewGitCheckoutBranch(gitSvc),
tools.NewGitAdd(gitSvc),
tools.NewGitCommit(gitSvc),
tools.NewGitPush(gitSvc),
)
}
if prSvc != nil {
tt = append(tt, tools.NewCreatePR(taskID, projectID, prSvc))
}
return tt
}
