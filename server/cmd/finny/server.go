package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/WorldOccupier/finny/server/internal/api"
	"github.com/WorldOccupier/finny/server/internal/database"
	"github.com/WorldOccupier/finny/server/internal/importservice"
)

const (
	DEFAULT_PORT      = 8080
	FINNY_PORT_ENV    = "FINNY_PORT"
	HEALTH_ROUTE      = "GET /health"
	JSON_CONTENT_TYPE = "application/json"
	HEALTH_STATUS     = "ok"
	SHUTDOWN_TIMEOUT  = 5 * time.Second
	FINNY_DB_PATH_ENV = "FINNY_DB_PATH"
	DEFAULT_DB_PATH   = "finny.db"
)

type server struct {
	logger *slog.Logger
}

func Run() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	db, err := database.Open(context.Background(), databasePath())
	if err != nil {
		logger.Error("database open failed", "error", err)
		return
	}
	defer db.Close()
	if err := database.Migrate(context.Background(), db); err != nil {
		logger.Error("database migration failed", "error", err)
		return
	}
	httpServer := newServerWithStore(serverPort(), logger, database.NewSQLiteStore(db))

	shutdownContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-shutdownContext.Done()

		ctx, cancel := context.WithTimeout(context.Background(), SHUTDOWN_TIMEOUT)
		defer cancel()

		if err := httpServer.Shutdown(ctx); err != nil {
			logger.Error("server shutdown failed", "error", err)
		}
	}()

	logger.Info("starting server", "address", httpServer.Addr)
	err = httpServer.ListenAndServe()

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}

func newServer(port int, logger *slog.Logger) *http.Server {
	return newServerWithStore(port, logger, nil)
}

func newServerWithStore(port int, logger *slog.Logger, store database.Store) *http.Server {
	app := &server{logger: logger}
	mux := http.NewServeMux()
	mux.HandleFunc(HEALTH_ROUTE, app.health)
	mux.Handle(api.DASHBOARD_ROUTE, api.NewDashboardHandler(store, logger))
	imports := api.NewImportHandler(importservice.New(store))
	mux.Handle(api.STATEMENTS_PREVIEW_ROUTE, imports)
	mux.Handle(api.STATEMENTS_CONFIRM_ROUTE, imports)
	mux.Handle(api.STATEMENTS_ROUTE, imports)
	mux.Handle(api.TRANSACTIONS_ROUTE, imports)
	mux.Handle(api.ACCOUNTS_ROUTE, imports)
	mux.Handle(api.SPENDING_SUMMARY_ROUTE, imports)

	return &http.Server{
		Addr:              ":" + strconv.Itoa(port),
		Handler:           app.loggingMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func databasePath() string {
	value := os.Getenv(FINNY_DB_PATH_ENV)
	if value == "" {
		return DEFAULT_DB_PATH
	}
	return value
}

func serverPort() int {
	value := os.Getenv(FINNY_PORT_ENV)
	if value == "" {
		return DEFAULT_PORT
	}

	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		return DEFAULT_PORT
	}

	return port
}

func (s *server) health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", JSON_CONTENT_TYPE)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": HEALTH_STATUS})
}

func (s *server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		s.logger.Info("request complete", "method", r.Method, "path", r.URL.Path, "duration", time.Since(started))
	})
}
