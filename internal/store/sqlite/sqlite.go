package sqlite

import (
"database/sql"
"fmt"

"github.com/google/uuid"
"github.com/jmoiron/sqlx"
"github.com/oskov/project-pipe/internal/store"

_ "modernc.org/sqlite"
)

type sqliteStore struct {
db           *sqlx.DB
projects     *projectRepo
tickets      *ticketRepo
tasks        *taskRepo
agentRuns    *agentRunRepo
agentMemory  *agentMemoryRepo
pullRequests *pullRequestRepository
}

// Ensure interface is satisfied at compile time.
var _ store.Store = (*sqliteStore)(nil)

// New opens a SQLite database and returns a Store.
func New(dsn string) (store.Store, error) {
db, err := sqlx.Open("sqlite", dsn)
if err != nil {
return nil, fmt.Errorf("open sqlite: %w", err)
}
db.SetMaxOpenConns(1)
if err := db.Ping(); err != nil {
return nil, fmt.Errorf("ping sqlite: %w", err)
}
s := &sqliteStore{db: db}
s.projects = &projectRepo{db: db}
s.tickets = &ticketRepo{db: db}
s.tasks = &taskRepo{db: db}
s.agentRuns = &agentRunRepo{db: db}
s.agentMemory = &agentMemoryRepo{db: db}
s.pullRequests = &pullRequestRepository{db: s}
return s, nil
}

func (s *sqliteStore) Projects() store.ProjectRepository           { return s.projects }
func (s *sqliteStore) Tickets() store.TicketRepository             { return s.tickets }
func (s *sqliteStore) Tasks() store.TaskRepository                 { return s.tasks }
func (s *sqliteStore) AgentRuns() store.AgentRunRepository         { return s.agentRuns }
func (s *sqliteStore) AgentMemory() store.AgentMemoryRepository    { return s.agentMemory }
func (s *sqliteStore) PullRequests() store.PullRequestRepository   { return s.pullRequests }
func (s *sqliteStore) Close() error                                { return s.db.Close() }

// RawDB exposes the underlying *sql.DB for operations like migrations.
func RawDB(s store.Store) *sql.DB {
return s.(*sqliteStore).db.DB
}

func newID() string { return uuid.New().String() }
