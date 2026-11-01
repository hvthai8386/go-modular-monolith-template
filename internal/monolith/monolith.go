// Package monolith hosts every module in one process: one gRPC server, one HTTP gateway in front
// of it, one database pool. Modules plug into it through Deps. Schema migrations are not run here;
// see `make migrate`.
package monolith

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/your-org/go-hexagon-template/internal/config"
)

const (
	shutdownTimeout   = 10 * time.Second
	readHeaderTimeout = 5 * time.Second
)

// Deps is everything a module may register itself against.
type Deps struct {
	GRPCServer grpc.ServiceRegistrar
	GatewayMux *runtime.ServeMux
	GRPCConn   *grpc.ClientConn
	DB         *pgxpool.Pool
	Logger     *slog.Logger
}

type Module interface {
	Name() string
	Register(ctx context.Context, deps Deps) error
}

type Monolith struct {
	cfg        config.Config
	db         *pgxpool.Pool
	grpcServer *grpc.Server
	grpcConn   *grpc.ClientConn
	httpServer *http.Server
}

func New(ctx context.Context, cfg config.Config, modules ...Module) (_ *Monolith, err error) {
	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database pool: %w", err)
	}
	defer func() {
		if err != nil {
			db.Close()
		}
	}()
	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(recoverUnary, logUnary))
	grpc_health_v1.RegisterHealthServer(grpcServer, health.NewServer())
	reflection.Register(grpcServer)

	// The gateway dials the gRPC listener over loopback so REST traffic passes the same interceptors
	// as native gRPC. A bare ":9090" listen address has no host to dial, so point it at localhost.
	host, port, err := net.SplitHostPort(cfg.GRPCAddr)
	if err != nil {
		return nil, fmt.Errorf("parse GRPC_ADDR %q: %w", cfg.GRPCAddr, err)
	}
	if host == "" {
		host = "localhost"
	}
	grpcConn, err := grpc.NewClient(net.JoinHostPort(host, port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("create gateway client: %w", err)
	}
	defer func() {
		if err != nil {
			_ = grpcConn.Close()
		}
	}()
	gatewayMux := runtime.NewServeMux()

	for _, module := range modules {
		deps := Deps{
			GRPCServer: grpcServer,
			GatewayMux: gatewayMux,
			GRPCConn:   grpcConn,
			DB:         db,
			Logger:     slog.Default().With("module", module.Name()),
		}
		if err := module.Register(ctx, deps); err != nil {
			return nil, fmt.Errorf("register module %q: %w", module.Name(), err)
		}
		slog.InfoContext(ctx, "module registered", "module", module.Name())
	}

	httpMux := http.NewServeMux()
	httpMux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	httpMux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(r.Context()); err != nil {
			http.Error(w, "database unreachable", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	httpMux.Handle("/", gatewayMux)

	return &Monolith{
		cfg:        cfg,
		db:         db,
		grpcServer: grpcServer,
		grpcConn:   grpcConn,
		httpServer: &http.Server{Addr: cfg.HTTPAddr, Handler: httpMux, ReadHeaderTimeout: readHeaderTimeout},
	}, nil
}

// Run serves until ctx is cancelled or a server fails, then drains connections and releases the pool.
func (m *Monolith) Run(ctx context.Context) error {
	grpcListener, err := net.Listen("tcp", m.cfg.GRPCAddr)
	if err != nil {
		return fmt.Errorf("listen on GRPC_ADDR %q: %w", m.cfg.GRPCAddr, err)
	}

	serveErr := make(chan error, 2)
	go func() {
		serveErr <- m.grpcServer.Serve(grpcListener)
	}()
	go func() {
		err := m.httpServer.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		serveErr <- err
	}()
	slog.InfoContext(ctx, "servers started", "grpc_addr", m.cfg.GRPCAddr, "http_addr", m.cfg.HTTPAddr)

	select {
	case <-ctx.Done():
		slog.InfoContext(ctx, "shutdown requested")
	case err := <-serveErr:
		if err != nil {
			m.stop(context.Background())
			return fmt.Errorf("server stopped: %w", err)
		}
	}
	m.stop(context.Background())
	return nil
}

func (m *Monolith) stop(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	// HTTP goes first so the gateway stops forwarding before the gRPC server drains.
	if err := m.httpServer.Shutdown(ctx); err != nil {
		slog.ErrorContext(ctx, "http server shutdown failed", "error", err)
	}

	drained := make(chan struct{})
	go func() {
		m.grpcServer.GracefulStop()
		close(drained)
	}()
	select {
	case <-drained:
	case <-ctx.Done():
		slog.WarnContext(ctx, "grpc drain timed out, stopping hard")
		m.grpcServer.Stop()
	}

	if err := m.grpcConn.Close(); err != nil {
		slog.ErrorContext(ctx, "gateway client close failed", "error", err)
	}
	m.db.Close()
	slog.InfoContext(ctx, "servers stopped")
}
