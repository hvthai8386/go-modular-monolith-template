package grpc

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	versioningv1 "github.com/your-org/go-hexagon-template/gen/versioning/v1"
	"github.com/your-org/go-hexagon-template/internal/versioning/core/domain"
	"github.com/your-org/go-hexagon-template/internal/versioning/core/port/in/inmock"
)

func TestVersioningHandler_GetVersion(t *testing.T) {
	service := &inmock.VersionService{
		GetVersionFn: func(context.Context) (domain.Version, error) {
			return domain.Version{App: "1.2.3", Go: "go1.26", Database: "PostgreSQL 17", Schema: 2}, nil
		},
	}
	handler := NewVersioningHandler(service, slog.Default())

	resp, err := handler.GetVersion(context.Background(), &versioningv1.GetVersionRequest{})
	if err != nil {
		t.Fatalf("GetVersion returned error: %v", err)
	}

	if resp.GetAppVersion() != "1.2.3" || resp.GetGoVersion() != "go1.26" ||
		resp.GetDatabaseVersion() != "PostgreSQL 17" || resp.GetSchemaVersion() != 2 {
		t.Errorf("GetVersion response = %v, fields do not match the domain version", resp)
	}
}

func TestVersioningHandler_GetVersion_ServiceError(t *testing.T) {
	service := &inmock.VersionService{
		GetVersionFn: func(context.Context) (domain.Version, error) {
			return domain.Version{}, errors.New("database down")
		},
	}
	handler := NewVersioningHandler(service, slog.Default())

	_, err := handler.GetVersion(context.Background(), &versioningv1.GetVersionRequest{})
	if status.Code(err) != codes.Internal {
		t.Errorf("GetVersion error code = %v, want %v", status.Code(err), codes.Internal)
	}
	if status.Convert(err).Message() != "failed to read version" {
		t.Errorf("GetVersion error message = %q leaks the cause", status.Convert(err).Message())
	}
}
