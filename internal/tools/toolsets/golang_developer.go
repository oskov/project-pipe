package toolsets

import (
	"github.com/oskov/project-pipe/internal/store"
	"github.com/oskov/project-pipe/internal/tools"
)

// GolangDeveloperTools returns the full set of tools for the Go developer agent.
// workDir must be the path to the cloned repository; filesystem and go toolchain
// tools are only included when workDir is non-empty.
func GolangDeveloperTools(
	workDir string,
	memoryRepo store.AgentMemoryRepository,
	projectID string,
) []tools.Tool {
	tt := []tools.Tool{
		tools.NewMemorySave(memoryRepo, projectID, store.AgentTypeGolangDeveloper),
		tools.NewMemoryGet(memoryRepo, projectID, store.AgentTypeGolangDeveloper),
		tools.NewMemoryList(memoryRepo, projectID, store.AgentTypeGolangDeveloper),
	}
	if workDir != "" {
		tt = append(tt,
			tools.NewReadFile(workDir),
			tools.NewReadFileRange(workDir),
			tools.NewWriteFile(workDir),
			tools.NewListFiles(workDir),
			tools.NewSearchCode(workDir),
			tools.NewGoCommand(workDir),
			tools.NewGoDefinitions(workDir),
			tools.NewGoReadDefinition(workDir),
		)
	}
	return tt
}
