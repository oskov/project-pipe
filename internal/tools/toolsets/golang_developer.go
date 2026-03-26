package toolsets

import (
	"github.com/oskov/project-pipe/internal/service"
	"github.com/oskov/project-pipe/internal/store"
	"github.com/oskov/project-pipe/internal/tools"
)

// GolangDeveloperTools returns the full set of tools for the Go developer agent.
// fsSvc, goSvc, parseSvc, and gitSvc are workspace-scoped; pass nil when workDir is not yet available.
func GolangDeveloperTools(
	memorySvc service.MemoryService,
	projectID string,
	fsSvc service.FilesystemService,
	goSvc service.GoToolchainService,
	parseSvc service.GoParseService,
	gitSvc service.GitService,
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
	return tt
}
