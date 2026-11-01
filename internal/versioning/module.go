// Package versioning reports the running versions of the app, Go, the database and its schema.
package versioning

import (
	"context"
	"fmt"

	versioningv1 "github.com/your-org/go-hexagon-template/gen/versioning/v1"
	"github.com/your-org/go-hexagon-template/internal/monolith"
	"github.com/your-org/go-hexagon-template/internal/versioning/adapter/in/grpc"
	"github.com/your-org/go-hexagon-template/internal/versioning/adapter/out/postgres"
	"github.com/your-org/go-hexagon-template/internal/versioning/core/domain"
	"github.com/your-org/go-hexagon-template/internal/versioning/core/usecase"
)

type Module struct {
	AppVersion string
}

var _ monolith.Module = (*Module)(nil)

func (m *Module) Name() string {
	return "versioning"
}

// Register wires the module from the outside in: driven adapter, use case, driving adapter.
func (m *Module) Register(ctx context.Context, deps monolith.Deps) error {
	appVersion, err := domain.NewAppVersion(m.AppVersion)
	if err != nil {
		return err
	}

	databaseInfo := postgres.NewDatabaseInfoReader(deps.DB)
	versionService := usecase.NewVersionService(appVersion, databaseInfo)

	handler := grpc.NewVersioningHandler(versionService, deps.Logger)
	versioningv1.RegisterVersioningServiceServer(deps.GRPCServer, handler)
	if err := versioningv1.RegisterVersioningServiceHandler(ctx, deps.GatewayMux, deps.GRPCConn); err != nil {
		return fmt.Errorf("register gateway handler: %w", err)
	}
	return nil
}
