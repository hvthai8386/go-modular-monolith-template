package grpc

import (
	"context"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	versioningv1 "github.com/hvthai8386/go-modular-monolith-template/gen/versioning/v1"
	"github.com/hvthai8386/go-modular-monolith-template/internal/versioning/core/port/in"
)

type VersioningHandler struct {
	versioningv1.UnimplementedVersioningServiceServer
	service in.VersionService
	logger  *slog.Logger
}

func NewVersioningHandler(service in.VersionService, logger *slog.Logger) *VersioningHandler {
	return &VersioningHandler{service: service, logger: logger}
}

func (h *VersioningHandler) GetVersion(ctx context.Context, _ *versioningv1.GetVersionRequest) (*versioningv1.GetVersionResponse, error) {
	version, err := h.service.GetVersion(ctx)
	if err != nil {
		// The cause stays in the log; clients only learn that the call failed.
		h.logger.ErrorContext(ctx, "get version failed", "error", err)
		return nil, status.Error(codes.Internal, "failed to read version")
	}
	return &versioningv1.GetVersionResponse{
		AppVersion:      version.App.String(),
		GoVersion:       version.Go,
		DatabaseVersion: version.Database,
		SchemaVersion:   version.Schema,
	}, nil
}
