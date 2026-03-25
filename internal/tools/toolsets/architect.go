package toolsets

import (
	"github.com/oskov/project-pipe/internal/store"
	"github.com/oskov/project-pipe/internal/tools"
)

// ArchitectTools returns the full set of tools for the architect agent.
// workDir is the path to the cloned repository (may be empty if not yet cloned).
func ArchitectTools(
	workDir string,
	memoryRepo store.AgentMemoryRepository,
	projectID string,
) []tools.Tool {
	tt := []tools.Tool{
		tools.NewMemorySave(memoryRepo, projectID, store.AgentTypeArchitect),
		tools.NewMemoryGet(memoryRepo, projectID, store.AgentTypeArchitect),
		tools.NewMemoryList(memoryRepo, projectID, store.AgentTypeArchitect),
	}
	if workDir != "" {
		tt = append(tt,
			tools.NewReadFile(workDir),
			tools.NewReadFileRange(workDir),
			tools.NewListFiles(workDir),
			tools.NewSearchCode(workDir),
			tools.NewGoDefinitions(workDir),
			tools.NewGoReadDefinition(workDir),
		)
	}
	return tt
}
