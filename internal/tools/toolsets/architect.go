package toolsets

import (
	"github.com/oskov/project-pipe/internal/service"
	"github.com/oskov/project-pipe/internal/store"
	"github.com/oskov/project-pipe/internal/tools"
)

// ArchitectTools returns the full set of tools for the architect agent.
// fsSvc and parseSvc are workspace-scoped; pass nil when workDir is not yet available.
func ArchitectTools(
	memorySvc service.MemoryService,
	projectID string,
	fsSvc service.FilesystemService,
	parseSvc service.GoParseService,
) []tools.Tool {
	tt := []tools.Tool{
		tools.NewMemorySave(memorySvc, projectID, store.AgentTypeArchitect),
		tools.NewMemoryGet(memorySvc, projectID, store.AgentTypeArchitect),
		tools.NewMemoryList(memorySvc, projectID, store.AgentTypeArchitect),
	}
	if fsSvc != nil {
		tt = append(tt,
			tools.NewReadFile(fsSvc),
			tools.NewReadFileRange(fsSvc),
			tools.NewListFiles(fsSvc),
			tools.NewSearchCode(fsSvc),
		)
	}
	if parseSvc != nil {
		tt = append(tt,
			tools.NewGoDefinitions(parseSvc),
			tools.NewGoReadDefinition(parseSvc),
		)
	}
	return tt
}
