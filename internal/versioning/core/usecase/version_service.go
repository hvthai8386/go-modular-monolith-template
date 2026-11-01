package usecase

import (
	"context"
	"fmt"
	"runtime"

	"github.com/hvthai8386/go-modular-monolith-template/internal/versioning/core/domain"
	"github.com/hvthai8386/go-modular-monolith-template/internal/versioning/core/port/in"
	"github.com/hvthai8386/go-modular-monolith-template/internal/versioning/core/port/out"
)

type VersionService struct {
	app      domain.AppVersion
	database out.DatabaseInfoReader
}

var _ in.VersionService = (*VersionService)(nil)

func NewVersionService(app domain.AppVersion, database out.DatabaseInfoReader) *VersionService {
	return &VersionService{app: app, database: database}
}

func (s *VersionService) GetVersion(ctx context.Context) (domain.Version, error) {
	serverVersion, err := s.database.ServerVersion(ctx)
	if err != nil {
		return domain.Version{}, fmt.Errorf("read database server version: %w", err)
	}

	schemaVersion, err := s.database.SchemaVersion(ctx)
	if err != nil {
		return domain.Version{}, fmt.Errorf("read database schema version: %w", err)
	}

	return domain.Version{
		App:      s.app,
		Go:       runtime.Version(),
		Database: serverVersion,
		Schema:   schemaVersion,
	}, nil
}
