package store

// Store aggregates all repositories.
type Store interface {
	Projects() ProjectRepository
	Tickets() TicketRepository
	Tasks() TaskRepository
	AgentRuns() AgentRunRepository
	AgentMemory() AgentMemoryRepository
	PullRequests() PullRequestRepository
	Close() error
}
