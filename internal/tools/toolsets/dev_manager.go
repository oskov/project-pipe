package toolsets

import (
	"github.com/oskov/project-pipe/internal/store"
	"github.com/oskov/project-pipe/internal/tools"
)

// DevManagerTools returns the full set of tools for the developer manager agent.
// Each runner parameter is a specialist developer agent that can be delegated to.
func DevManagerTools(
	memoryRepo store.AgentMemoryRepository,
	projectID string,
	golangDeveloperRunner tools.AgentRunner,
) []tools.Tool {
	return []tools.Tool{
		// project-scoped memory
		tools.NewMemorySave(memoryRepo, projectID, store.AgentTypeDevManager),
		tools.NewMemoryGet(memoryRepo, projectID, store.AgentTypeDevManager),
		tools.NewMemoryList(memoryRepo, projectID, store.AgentTypeDevManager),
		// specialist developers
		tools.NewRunAgent(
			"golang_developer",
			"Delegate Go implementation work to the Go developer agent. Provide the architectural plan, working directory, and expected deliverables. The agent will implement changes, run go build/test/vet, and return a summary.",
			projectID,
			golangDeveloperRunner,
		),
	}
}
