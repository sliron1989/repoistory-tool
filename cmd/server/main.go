package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sliron1989/repoistory-tool/internal/api"
	"github.com/sliron1989/repoistory-tool/internal/github"

	_ "github.com/sliron1989/repoistory-tool/docs"
)

// @title Legit - Repository Creation Platform API
// @version 1.0.0
// @description Platform microservice that automates GitHub repository creation with production-ready defaults.
// @host localhost:8080
// @BasePath /

const version = "1.0.0"

func main() {
	// Configure structured logging
	logFormat := os.Getenv("LOG_FORMAT")
	var handler slog.Handler
	if logFormat == "json" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	slog.SetDefault(slog.New(handler))

	// GitHub service
	token := os.Getenv("GITHUB_TOKEN")
	ghService, err := github.NewService(token)
	if err != nil {
		slog.Error("failed to initialize GitHub service", "error", err)
		os.Exit(1)
	}
	slog.Info("GitHub service initialized")

	// HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	h := api.NewHandler(ghService, version)
	router := api.NewRouter(h)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		slog.Info("shutting down server")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("graceful shutdown failed", "error", err)
		}
	}()

	slog.Info("server starting", "port", port, "version", version)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}
