package postgres

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Runs only against a real database that `make migrate` has already been applied to;
// set DATABASE_URL to enable it.
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

	reader := NewDatabaseInfoReader(pool)

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
		t.Errorf("SchemaVersion = %d, want at least 1; run `make migrate` first", schemaVersion)
	}
}
