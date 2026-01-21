package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ziyad/cms-ai/server/internal/api"
	"github.com/ziyad/cms-ai/server/internal/logger"
)

func main() {
	// Initialize structured logging
	logLevel := logger.LogLevel(env("LOG_LEVEL", "info"))
	logFormat := env("LOG_FORMAT", "json")

	logger.Initialize(&logger.Config{
		Level:  logLevel,
		Format: logFormat,
	})

	logger.Logger.Info("server_starting",
		"log_level", logLevel,
		"log_format", logFormat,
	)

	// Support both PORT (Railway) and ADDR (local dev)
	port := env("PORT", "")
	addr := env("ADDR", "")
	if port != "" {
		addr = ":" + port
	} else if addr == "" {
		addr = ":8080"
	}

	srv, worker := api.NewServerWithWorker()
	worker.Start()
	defer worker.Stop()

	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Logger.Info("server_listening", "addr", addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Error("server_error", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Logger.Info("server_shutting_down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(ctx); err != nil {
		logger.Logger.Error("shutdown_error", "error", err)
	} else {
		logger.Logger.Info("server_shutdown_complete")
	}
}

func env(key string, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
