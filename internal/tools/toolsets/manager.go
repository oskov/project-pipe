package toolsets

import (
	"github.com/oskov/project-pipe/internal/store"
	"github.com/oskov/project-pipe/internal/tools"
)

// ManagerTools returns the full set of tools for the manager agent.
// devManagerRunner allows the manager to delegate implementation work
// to the developer manager, which in turn routes to specialist developers.
func ManagerTools(
	ticketRepo store.TicketRepository,
	memoryRepo store.AgentMemoryRepository,
	projectID string,
	architectRunner tools.AgentRunner,
	devManagerRunner tools.AgentRunner,
) []tools.Tool {
	return []tools.Tool{
		// project management
		tools.NewListTickets(ticketRepo, projectID),
		tools.NewCreateTicket(ticketRepo, projectID),
		tools.NewGetTicket(ticketRepo),
		// project-scoped memory
		tools.NewMemorySave(memoryRepo, projectID, store.AgentTypeManager),
		tools.NewMemoryGet(memoryRepo, projectID, store.AgentTypeManager),
		tools.NewMemoryList(memoryRepo, projectID, store.AgentTypeManager),
		// sub-agent delegation
		tools.NewRunAgent(
			"architect",
			"Delegate to the architect agent to produce a detailed architectural change plan. Provide full context: ticket details, current codebase state, and requirements. Returns a structured plan to pass to the developer manager.",
			projectID,
			architectRunner,
		),
		tools.NewRunAgent(
			"dev_manager",
			"Delegate implementation to the developer manager agent. Provide the full architectural plan from the architect. The dev manager will route work to the appropriate specialist developers and return an implementation summary.",
			projectID,
			devManagerRunner,
		),
	}
}
