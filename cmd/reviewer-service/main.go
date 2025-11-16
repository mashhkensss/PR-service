package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mashhkensss/PR-service/internal/app"
	"github.com/mashhkensss/PR-service/internal/config"
)

const startupTimeout = 10 * time.Second

func run() int {
	rootLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: false})).With("app", "reviewer-service")
	logger := rootLogger.With("component", "app")

	cfg, err := config.Load()
	if err != nil {
		logger.Error("config load failed", "error", err)
		return 1
	}
	logger.Info("config loaded",
		"http_addr", cfg.HTTP.Addr,
		"read_timeout", cfg.HTTP.ReadTimeout,
		"write_timeout", cfg.HTTP.WriteTimeout,
		"shutdown_timeout", cfg.HTTP.ShutdownTimeout,
	)

	startCtx, cancel := context.WithTimeout(context.Background(), startupTimeout)
	defer cancel()

	handler, closeApp, err := app.Build(startCtx, cfg, rootLogger)
	if err != nil {
		logger.Error("bootstrap failed", "error", err)
		return 1
	}
	defer func() {
		if err := closeApp(); err != nil {
			logger.Error("cleanup failed", "error", err)
		}
	}()

	server := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           handler,
		ReadHeaderTimeout: cfg.HTTP.ReadTimeout,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("http listen start", "addr", cfg.HTTP.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		logger.Info("shutdown signal received", "signal", sig.String())
	case err := <-serverErr:
		if err != nil {
			logger.Error("server failed", "error", err)
			return 1
		}
		logger.Info("server stopped")
		return 0
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown failed", "error", err)
		return 1
	}

	if err := <-serverErr; err != nil {
		logger.Error("server failed", "error", err)
		return 1
	}

	logger.Info("server shutdown complete")
	return 0
}

func main() {
	os.Exit(run())
}
