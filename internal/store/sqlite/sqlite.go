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
db          *sqlx.DB
projects    *projectRepo
tickets     *ticketRepo
tasks       *taskRepo
agentRuns   *agentRunRepo
agentMemory *agentMemoryRepo
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
return &sqliteStore{
db:          db,
projects:    &projectRepo{db: db},
tickets:     &ticketRepo{db: db},
tasks:       &taskRepo{db: db},
agentRuns:   &agentRunRepo{db: db},
agentMemory: &agentMemoryRepo{db: db},
}, nil
}

func (s *sqliteStore) Projects() store.ProjectRepository        { return s.projects }
func (s *sqliteStore) Tickets() store.TicketRepository          { return s.tickets }
func (s *sqliteStore) Tasks() store.TaskRepository              { return s.tasks }
func (s *sqliteStore) AgentRuns() store.AgentRunRepository      { return s.agentRuns }
func (s *sqliteStore) AgentMemory() store.AgentMemoryRepository { return s.agentMemory }
func (s *sqliteStore) Close() error                             { return s.db.Close() }

// RawDB exposes the underlying *sql.DB for operations like migrations.
func RawDB(s store.Store) *sql.DB {
return s.(*sqliteStore).db.DB
}

func newID() string { return uuid.New().String() }
