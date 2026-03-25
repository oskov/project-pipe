package worker

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/oskov/project-pipe/internal/service"
)

const pollInterval = 2 * time.Second

// Dispatcher manages a pool of worker goroutines that claim and execute tasks.
type Dispatcher struct {
	tasks  service.TaskService
	logDir string
	size   int
	wg     sync.WaitGroup
}

// New creates a Dispatcher with the given worker pool size and log directory.
func New(tasks service.TaskService, logDir string, size int) *Dispatcher {
	return &Dispatcher{tasks: tasks, logDir: logDir, size: size}
}

// Start spawns worker goroutines. They run until ctx is cancelled.
// Call Wait to block until all workers have exited.
func (d *Dispatcher) Start(ctx context.Context) {
	if err := os.MkdirAll(d.logDir, 0o755); err != nil {
		slog.Error("worker: cannot create log directory", "dir", d.logDir, "error", err)
	}

	for i := range d.size {
		d.wg.Add(1)
		go d.runWorker(ctx, i)
	}
}

// Wait blocks until all worker goroutines have stopped.
func (d *Dispatcher) Wait() { d.wg.Wait() }

func (d *Dispatcher) runWorker(ctx context.Context, id int) {
	defer d.wg.Done()
	slog.Info("worker started", "id", id)

	for {
		select {
		case <-ctx.Done():
			slog.Info("worker stopped", "id", id)
			return
		default:
		}

		task, err := d.tasks.ClaimNext(ctx)
		if err != nil {
			slog.Error("worker: claim error", "worker_id", id, "error", err)
			d.sleep(ctx)
			continue
		}
		if task == nil {
			// No claimable task — wait before polling again.
			d.sleep(ctx)
			continue
		}

		logger, closeLog := openTaskLogger(d.logDir, task.ID)
		slog.Info("worker: executing task", "worker_id", id, "task_id", task.ID, "project_id", task.ProjectID)

		if err := d.tasks.Execute(ctx, task, logger); err != nil {
			slog.Error("worker: task failed", "worker_id", id, "task_id", task.ID, "error", err)
		}

		closeLog()
	}
}

// sleep waits for pollInterval or until ctx is cancelled.
func (d *Dispatcher) sleep(ctx context.Context) {
	select {
	case <-ctx.Done():
	case <-time.After(pollInterval):
	}
}

// openTaskLogger creates a slog.Logger writing to {logDir}/{taskID}.log.
// The returned closer flushes and closes the underlying file.
func openTaskLogger(logDir, taskID string) (*slog.Logger, func()) {
	path := filepath.Join(logDir, taskID+".log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		slog.Warn("worker: cannot open task log file, using default logger", "path", path, "error", err)
		return slog.Default(), func() {}
	}
	logger := slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{Level: slog.LevelDebug}))
	return logger, func() { _ = f.Close() }
}
