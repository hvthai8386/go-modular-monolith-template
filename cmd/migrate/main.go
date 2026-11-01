// Command migrate applies the embedded goose migrations to the database named by DATABASE_URL.
// It is a separate binary so deployments decide when schema changes run, and the app itself
// needs no DDL privileges.
//
// Usage: migrate [up|down|status|version]. The default is up.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" database/sql driver
	"github.com/pressly/goose/v3"

	"github.com/your-org/go-hexagon-template/internal/config"
	"github.com/your-org/go-hexagon-template/internal/logging"
	"github.com/your-org/go-hexagon-template/migrations"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fatal("failed to load config", err)
	}
	logging.Setup(cfg.LogLevel, cfg.Env == "production", "go-hexagon-template-migrate")

	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		fatal("failed to open database", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.ErrorContext(context.Background(), "failed to close database", slog.Any("error", err))
		}
	}()

	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations.FS)
	if err != nil {
		fatal("failed to create goose provider", err)
	}

	ctx := context.Background()
	switch command {
	case "up":
		results, err := provider.Up(ctx)
		if err != nil {
			fatal("migration up failed", err)
		}
		for _, result := range results {
			slog.InfoContext(ctx, "migration applied", "version", result.Source.Version, "path", result.Source.Path, "duration_ms", result.Duration.Milliseconds())
		}
		slog.InfoContext(ctx, "migrations up to date", "applied", len(results))

	case "down":
		result, err := provider.Down(ctx)
		if err != nil {
			fatal("migration down failed", err)
		}
		slog.InfoContext(ctx, "migration rolled back", "version", result.Source.Version, "path", result.Source.Path)

	case "status":
		statuses, err := provider.Status(ctx)
		if err != nil {
			fatal("migration status failed", err)
		}
		for _, status := range statuses {
			fmt.Printf("%-8s %5d  %s\n", status.State, status.Source.Version, status.Source.Path)
		}

	case "version":
		version, err := provider.GetDBVersion(ctx)
		if err != nil {
			fatal("migration version failed", err)
		}
		fmt.Println(version)

	default:
		fatal("unknown command", fmt.Errorf("%q is not one of up, down, status, version", command))
	}
}

// fatal logs at Error, which every level setting emits, and exits.
func fatal(msg string, err error) {
	slog.ErrorContext(context.Background(), msg, slog.Any("error", err))
	os.Exit(1)
}
