package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/your-org/go-hexagon-template/internal/versioning/adapter/out/postgres/sqlcgen"
	"github.com/your-org/go-hexagon-template/internal/versioning/core/port/out"
)

type DatabaseInfoReader struct {
	queries *sqlcgen.Queries
}

var _ out.DatabaseInfoReader = (*DatabaseInfoReader)(nil)

func NewDatabaseInfoReader(pool *pgxpool.Pool) *DatabaseInfoReader {
	return &DatabaseInfoReader{queries: sqlcgen.New(pool)}
}

func (r *DatabaseInfoReader) ServerVersion(ctx context.Context) (string, error) {
	return r.queries.ServerVersion(ctx)
}

func (r *DatabaseInfoReader) SchemaVersion(ctx context.Context) (int64, error) {
	return r.queries.SchemaVersion(ctx)
}
