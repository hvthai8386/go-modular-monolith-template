package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hvthai8386/go-modular-monolith-template/internal/db"
	"github.com/hvthai8386/go-modular-monolith-template/internal/versioning/core/port/out"
)

type DatabaseInfoReader struct {
	queries *db.Queries
}

var _ out.DatabaseInfoReader = (*DatabaseInfoReader)(nil)

func NewDatabaseInfoReader(pool *pgxpool.Pool) *DatabaseInfoReader {
	return &DatabaseInfoReader{queries: db.New(pool)}
}

func (r *DatabaseInfoReader) ServerVersion(ctx context.Context) (string, error) {
	return r.queries.ServerVersion(ctx)
}

func (r *DatabaseInfoReader) SchemaVersion(ctx context.Context) (int64, error) {
	return r.queries.SchemaVersion(ctx)
}
