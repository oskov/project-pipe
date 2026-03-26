package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oskov/project-pipe/internal/api"
	"github.com/oskov/project-pipe/internal/config"
	"github.com/oskov/project-pipe/internal/factory"
	"github.com/oskov/project-pipe/internal/llm/langchain"
	applogger "github.com/oskov/project-pipe/internal/logger"
	"github.com/oskov/project-pipe/internal/migrate"
	"github.com/oskov/project-pipe/internal/service"
	sqlitestore "github.com/oskov/project-pipe/internal/store/sqlite"
	"github.com/oskov/project-pipe/internal/worker"
)

const workerShutdownTimeout = 5 * time.Minute

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "config.yaml"
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := applogger.New(cfg.Server.LogLevel, cfg.Server.JSONLog)
	slog.SetDefault(logger)

	// ── database ────────────────────────────────────────────────────────────
	db, err := sqlitestore.New(cfg.DB.DSN)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer db.Close()

	if err := migrate.Run(context.Background(), sqlitestore.RawDB(db), "migrations"); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	// ── LLM ─────────────────────────────────────────────────────────────────
	llmClient, err := langchain.New(cfg.LLM)
	if err != nil {
		return fmt.Errorf("init llm: %w", err)
	}

	// ── services ─────────────────────────────────────────────────────────────
	taskSvc := service.NewTaskService(db.Tasks(), db.Projects(), factory.ManagerFactory(db, llmClient))

	// ── HTTP server ──────────────────────────────────────────────────────────
	router := api.NewRouter(db, llmClient, taskSvc, logger)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ── worker pool ──────────────────────────────────────────────────────────
	dispatcher := worker.New(taskSvc, cfg.Worker.LogDir, cfg.Worker.PoolSize)

	workerCtx, cancelWorkers := context.WithCancel(context.Background())
	dispatcher.Start(workerCtx)

	// ── start + graceful shutdown ────────────────────────────────────────────
	errCh := make(chan error, 1)
	go func() {
		logger.Info("server started", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		cancelWorkers()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), workerShutdownTimeout)
		defer shutdownCancel()
		dispatcher.Wait(shutdownCtx)
		return fmt.Errorf("server error: %w", err)
	case sig := <-quit:
		logger.Info("shutting down", "signal", sig)
	}

	cancelWorkers()
	logger.Info("waiting for running tasks to finish...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), workerShutdownTimeout)
	defer shutdownCancel()
	dispatcher.Wait(shutdownCtx)
	logger.Info("all tasks finished, stopping http server")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}
