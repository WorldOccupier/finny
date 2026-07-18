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
)

const (
	DEFAULT_PORT      = 8080
	FINNY_PORT_ENV    = "FINNY_PORT"
	HEALTH_ROUTE      = "GET /health"
	JSON_CONTENT_TYPE = "application/json"
	HEALTH_STATUS     = "ok"
	SHUTDOWN_TIMEOUT  = 5 * time.Second
)

type server struct {
	logger *slog.Logger
}

func Run() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	httpServer := newServer(serverPort(), logger)

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
	err := httpServer.ListenAndServe()

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}

func newServer(port int, logger *slog.Logger) *http.Server {
	app := &server{logger: logger}
	mux := http.NewServeMux()
	mux.HandleFunc(HEALTH_ROUTE, app.health)

	return &http.Server{
		Addr:              ":" + strconv.Itoa(port),
		Handler:           app.loggingMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}
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
