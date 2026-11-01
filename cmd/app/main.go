package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/hvthai8386/go-modular-monolith-template/internal/config"
	"github.com/hvthai8386/go-modular-monolith-template/internal/logging"
	"github.com/hvthai8386/go-modular-monolith-template/internal/monolith"
	"github.com/hvthai8386/go-modular-monolith-template/internal/versioning"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fatal("failed to load config", err)
	}
	logging.Setup(cfg.LogLevel, cfg.Env == "production", "go-modular-monolith-template")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	// After the first signal starts the drain, hand the signals back to Go's
	// default handling so a second Ctrl-C ends the process at once instead of
	// being swallowed until the drain finishes.
	context.AfterFunc(ctx, stop)

	app, err := monolith.New(ctx, cfg,
		&versioning.Module{AppVersion: cfg.AppVersion},
	)
	if err != nil {
		fatal("failed to start", err)
	}
	if err := app.Run(ctx); err != nil {
		fatal("server stopped", err)
	}
}

// fatal logs at Error, which every level setting emits, and exits.
func fatal(msg string, err error) {
	slog.ErrorContext(context.Background(), msg, slog.Any("error", err))
	os.Exit(1)
}
