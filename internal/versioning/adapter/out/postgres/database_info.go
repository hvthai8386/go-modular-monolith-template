package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/your-org/go-hexagon-template/internal/versioning/adapter/out/postgres/sqlcgen"
	"github.com/your-org/go-hexagon-template/internal/versioning/core/port/out"
	"github.com/your-org/go-hexagon-template/migrations"
)

type DatabaseInfoReader struct {
	queries *sqlcgen.Queries
	goose   *goose.Provider
}

var _ out.DatabaseInfoReader = (*DatabaseInfoReader)(nil)

func NewDatabaseInfoReader(pool *pgxpool.Pool) (*DatabaseInfoReader, error) {
	// goose works on database/sql, so it gets a *sql.DB view of the pool. Closing that view would
	// not close the pool, so nothing here needs a Close method.
	gooseProvider, err := goose.NewProvider(goose.DialectPostgres, stdlib.OpenDBFromPool(pool), migrations.FS)
	if err != nil {
		return nil, fmt.Errorf("create goose provider: %w", err)
	}
	return &DatabaseInfoReader{queries: sqlcgen.New(pool), goose: gooseProvider}, nil
}

func (r *DatabaseInfoReader) ServerVersion(ctx context.Context) (string, error) {
	return r.queries.ServerVersion(ctx)
}

func (r *DatabaseInfoReader) SchemaVersion(ctx context.Context) (int64, error) {
	return r.goose.GetDBVersion(ctx)
}
