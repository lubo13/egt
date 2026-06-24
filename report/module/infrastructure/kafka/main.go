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

	"git.infra.egt.com/report/module/infrastructure"
	"git.infra.egt.com/report/module/infrastructure/database"
	"git.infra.egt.com/report/module/infrastructure/database/repository"
	"git.infra.egt.com/report/module/infrastructure/kafka/processor"
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
			func() context.Context { return context.Background() },
			zap.NewExample,
			// Worker
			NewWorker,
			// Event consumer
			NewDeviceEventConsumer,
			func(deviceEventConsumer *DeviceEventConsumer) []ConsumerInterface {
				return []ConsumerInterface{deviceEventConsumer}
			},
			fx.Annotate(
				processor.NewDeviceEventProcessor,
				fx.As(new(processor.MessageProcessorInterface)),
			),
			// Database
			database.NewDatabasePool,
			// Repositories
			fx.Annotate(
				database.NewTransactionManager,
				fx.As(new(processor.TransactionManagerInterface)),
				fx.As(new(repository.QueryExecutorInterface)),
			),
			fx.Annotate(
				repository.NewIdempotentIdRepository,
				fx.As(new(processor.IdempotentIdRepositoryInterface)),
			),
			fx.Annotate(
				repository.NewDeviceEventRepository,
				fx.As(new(application.DeviceEventSaverRepositoryInterface)),
			),
			// Handler
			fx.Annotate(
				application.NewDeviceEventHandler,
				fx.As(new(processor.DeviceEventHandlerInterface)),
			),
			infrastructure.NewConfig,
			grpcserver.NewGrpcServer,
			grpcserver.NewGrpcHealthServer,
		),
		fx.Decorate(
			func(
				config *infrastructure.Config,
				logger *zap.Logger,
				idempotentIdRepository processor.IdempotentIdRepositoryInterface,
				original processor.MessageProcessorInterface,
				transactionManager processor.TransactionManagerInterface,
			) processor.MessageProcessorInterface {
				return processor.NewIdempotentProcessorDecorator(
					config,
					logger,
					idempotentIdRepository,
					original,
					transactionManager,
				)
			},
		),
		fx.Invoke(
			func(lc fx.Lifecycle, pool *pgxpool.Pool) {
				infrastructure.RegisterDatabaseLifecycleHooks(lc, pool)
			},
		),
		fx.Invoke(func(
			ctx context.Context,
			lc fx.Lifecycle,
			worker *Wroker,
			logger *zap.Logger,
		) {
			addHook(ctx, lc, worker, logger)
		}),
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

func addHook(
	ctx context.Context,
	lc fx.Lifecycle,
	worker *Wroker,
	logger *zap.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Srart background consumer worker...")

			worker.Run(ctx)

			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping background worker...")
			done := make(chan struct{})
			go func() {
				worker.StopProcessingGracefully()
				close(done)
			}()

			select {
			case <-done:
				logger.Info("Background workers have stopped")

				return nil
			case <-time.After(5 * time.Second):
				logger.Warn("Background workers do not stop in time")

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
			grpcserver.GrpcServe(logger, config, grpcServer)

			grpcserver.CheckCriticalDependencies(ctx, healthServer, logger, pool)

			return nil
		},
	})
}
