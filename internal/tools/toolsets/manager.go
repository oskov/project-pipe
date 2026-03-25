package toolsets

import (
	"github.com/oskov/project-pipe/internal/service"
	"github.com/oskov/project-pipe/internal/store"
	"github.com/oskov/project-pipe/internal/tools"
)

// ManagerTools returns the full set of tools for the manager agent.
func ManagerTools(
	ticketSvc service.TicketService,
	memorySvc service.MemoryService,
	projectID string,
	architectRunner tools.AgentRunner,
	devManagerRunner tools.AgentRunner,
) []tools.Tool {
	return []tools.Tool{
		// project management
		tools.NewListTickets(ticketSvc, projectID),
		tools.NewCreateTicket(ticketSvc, projectID),
		tools.NewGetTicket(ticketSvc),
		// project-scoped memory
		tools.NewMemorySave(memorySvc, projectID, store.AgentTypeManager),
		tools.NewMemoryGet(memorySvc, projectID, store.AgentTypeManager),
		tools.NewMemoryList(memorySvc, projectID, store.AgentTypeManager),
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
