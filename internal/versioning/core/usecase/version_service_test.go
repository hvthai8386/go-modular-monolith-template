package usecase

import (
	"context"
	"errors"
	"runtime"
	"testing"

	"github.com/your-org/go-hexagon-template/internal/versioning/core/domain"
	"github.com/your-org/go-hexagon-template/internal/versioning/core/port/out/outmock"
)

func TestVersionService_GetVersion(t *testing.T) {
	database := &outmock.DatabaseInfoReader{
		ServerVersionFn: func(context.Context) (string, error) { return "PostgreSQL 17.2", nil },
		SchemaVersionFn: func(context.Context) (int64, error) { return 3, nil },
	}
	service := NewVersionService(domain.AppVersion("1.2.3"), database)

	version, err := service.GetVersion(context.Background())
	if err != nil {
		t.Fatalf("GetVersion returned error: %v", err)
	}

	want := domain.Version{App: "1.2.3", Go: runtime.Version(), Database: "PostgreSQL 17.2", Schema: 3}
	if version != want {
		t.Errorf("GetVersion = %+v, want %+v", version, want)
	}
}

func TestVersionService_GetVersion_DatabaseError(t *testing.T) {
	databaseErr := errors.New("connection refused")
	database := &outmock.DatabaseInfoReader{
		ServerVersionFn: func(context.Context) (string, error) { return "", databaseErr },
	}
	service := NewVersionService(domain.AppVersion("1.2.3"), database)

	_, err := service.GetVersion(context.Background())
	if !errors.Is(err, databaseErr) {
		t.Errorf("GetVersion error = %v, want it to wrap %v", err, databaseErr)
	}
}
