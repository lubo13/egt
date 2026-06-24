package main

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"

	"git.infra.egt.com/report/module/infrastructure"
	"git.infra.egt.com/report/module/infrastructure/database"
	"git.infra.egt.com/report/module/infrastructure/database/repository"
	"git.infra.egt.com/report/module/infrastructure/port"
	grpcserver "git.infra.egt.com/report/module/infrastructure/server/grpc"
	"git.infra.egt.com/report/module/internal/application"
)

// TODO: DI framework init could be optimized
func main() {
	fx.New(
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
		fx.Provide(
			// Server
			func() context.Context { return context.Background() },
			zap.NewExample,
			// Database
			database.NewDatabasePool,
			fx.Annotate(
				database.NewTransactionManager,
				fx.As(new(repository.QueryExecutorInterface)),
			),
			// Repositories
			fx.Annotate(
				repository.NewDeviceEventRepository,
				fx.As(new(application.DeviceEventRepositoryInterface)),
			),
			infrastructure.NewConfig,
			grpcserver.NewGrpcServer,
			grpcserver.NewGrpcHealthServer,
			port.NewDeviceEventGPRCHandler,
			// Fetcher
			application.NewDeviceEventFetcher,
		),
		fx.Invoke(
			func(lc fx.Lifecycle, pool *pgxpool.Pool) {
				infrastructure.RegisterDatabaseLifecycleHooks(lc, pool)
			},
		),
		fx.Invoke(
			func(grpcServer *grpc.Server, deviceEventGPRCHandler *port.DeviceEventGPRCHandler) {
				port.RegisterDeviceEventGRPCHandler(grpcServer, deviceEventGPRCHandler)
			},
		),
		fx.Invoke(func(
			lc fx.Lifecycle,
			config *infrastructure.Config,
			grpcServer *grpc.Server,
			healthServer *health.Server,
			logger *zap.Logger,
			pool *pgxpool.Pool,
		) {
			addGRPCHook(lc, config, grpcServer, healthServer, logger, pool)
		}),
	).Run()
}

func addGRPCHook(
	lc fx.Lifecycle,
	config *infrastructure.Config,
	grpcServer *grpc.Server,
	healthServer *health.Server,
	logger *zap.Logger,
	pool *pgxpool.Pool,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			grpcserver.GrpcServe(logger, config, grpcServer)

			grpcserver.CheckCriticalDependencies(ctx, healthServer, logger, pool)

			return nil
		},
	})
}
