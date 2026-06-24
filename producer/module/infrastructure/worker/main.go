package main

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"

	"git.infra.egt.com/producer/module/infrastructure"
	"git.infra.egt.com/producer/module/infrastructure/database"
	"git.infra.egt.com/producer/module/infrastructure/database/repository"
	"git.infra.egt.com/producer/module/infrastructure/kafka"
	"git.infra.egt.com/producer/module/infrastructure/server"
)

func main() {
	fx.New(
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
		fx.Provide(
			func() context.Context { return context.Background() },
			zap.NewExample,
			// Event dispatcher
			kafka.NewEventDispather,
			// Repositories
			fx.Annotate(
				repository.NewMessageRepository,
				fx.As(new(kafka.MessageRepositoryInterface)),
			),
			// Database
			database.NewDatabasePool,
			fx.Annotate(
				database.NewTransactionManager,
				fx.As(new(repository.QueryExecutor)),
				fx.As(new(kafka.TransactionManagerInterface)),
			),
			infrastructure.NewConfig,
			server.NewGrpcServer,
			server.NewGrpcHealthServer,
		),
		fx.Invoke(
			func(lc fx.Lifecycle, log *zap.Logger, eventDispatcher *kafka.EventDispather) {
				addHook(lc, log, eventDispatcher)
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

func addHook(lc fx.Lifecycle, log *zap.Logger, eventDispatcher *kafka.EventDispather) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			log.Info("Stopping background worker...")
			done := make(chan struct{})
			go func() {
				eventDispatcher.StopProcessingGracefully()
				close(done)
			}()

			select {
			case <-done:
				log.Info("Background workers have stopped")

				return nil
			case <-time.After(5 * time.Second):
				log.Warn("Background workers do not stop in time")

				return nil
			}
		},
	})
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
			server.GrpcServe(logger, config, grpcServer)

			server.CheckCriticalDependencies(ctx, healthServer, logger, pool)

			return nil
		},
	})
}
