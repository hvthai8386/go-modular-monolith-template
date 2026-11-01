package postgres

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/your-org/go-hexagon-template/migrations"
)

// Runs only against a real database; set DATABASE_URL to enable it.
func TestDatabaseInfoReader(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set")
	}
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open pool: %v", err)
	}
	defer pool.Close()

	migrator, err := goose.NewProvider(goose.DialectPostgres, stdlib.OpenDBFromPool(pool), migrations.FS)
	if err != nil {
		t.Fatalf("create goose provider: %v", err)
	}
	if _, err := migrator.Up(ctx); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	reader, err := NewDatabaseInfoReader(pool)
	if err != nil {
		t.Fatalf("NewDatabaseInfoReader: %v", err)
	}

	serverVersion, err := reader.ServerVersion(ctx)
	if err != nil {
		t.Fatalf("ServerVersion: %v", err)
	}
	if !strings.HasPrefix(serverVersion, "PostgreSQL") {
		t.Errorf("ServerVersion = %q, want a PostgreSQL version string", serverVersion)
	}

	schemaVersion, err := reader.SchemaVersion(ctx)
	if err != nil {
		t.Fatalf("SchemaVersion: %v", err)
	}
	if schemaVersion < 1 {
		t.Errorf("SchemaVersion = %d, want at least 1 after migrations", schemaVersion)
	}
}
