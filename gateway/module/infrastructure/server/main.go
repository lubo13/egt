package main

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"

	"git.infra.egt.com/gateway/module/infrastructure"
	"git.infra.egt.com/gateway/module/infrastructure/database"
	"git.infra.egt.com/gateway/module/infrastructure/database/repository"
	"git.infra.egt.com/gateway/module/infrastructure/kafka"
	"git.infra.egt.com/gateway/module/infrastructure/port"
	"git.infra.egt.com/gateway/module/internal/application/service"
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
			NewHTTPServer,
			fx.Annotate(
				port.NewDeviceEventHandler,
				fx.As(new(Route)),
			),
			func(route Route) []Route {
				return []Route{route}
			},
			zap.NewExample,
			NewServeMux,
			// Route
			fx.Annotate(
				service.NewDeviceEventService,
				fx.As(new(port.DeviceEventProcessorInterface)),
			),
			// Event dispatcher
			fx.Annotate(
				kafka.NewEventDispather,
				fx.As(new(repository.DomainEventProducerInterface)),
			),
			// Repositories
			fx.Annotate(
				repository.NewDeviceEventRepository,
				fx.As(new(service.DeviceEventRepositoryInterface)),
				fx.As(new(repository.ModelRepositoryInterface)),
			),
			// Database
			database.NewDatabasePool,
			fx.Annotate(
				database.NewTransactionManager,
				fx.As(new(service.TransactionManagerInterface)),
				fx.As(new(repository.QueryExecutor)),
			),
			fx.Annotate(
				repository.NewMessageRepository,
				fx.As(new(kafka.MessageRepositoryInterface)),
			),
			infrastructure.NewConfig,
			NewGrpcServer,
			NewGrpcHealthServer,
		),
		fx.Decorate(
			func(
				logger *zap.Logger,
				original service.DeviceEventRepositoryInterface,
				domainEventProducer repository.DomainEventProducerInterface,
			) service.DeviceEventRepositoryInterface {
				return repository.NewDomainEventRepositoryDecorator(
					logger,
					original,
					domainEventProducer,
				)
			},
		),
		fx.Invoke(
			func(lc fx.Lifecycle, log *zap.Logger, deviceEventService port.DeviceEventProcessorInterface) {
				addHook(lc, deviceEventService, log)
			},
		),
		fx.Invoke(
			func(lc fx.Lifecycle, pool *pgxpool.Pool) {
				infrastructure.RegisterDatabaseLifecycleHooks(lc, pool)
			},
		),
		fx.Invoke(func(*http.Server) {}),
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
	lc fx.Lifecycle,
	deviceEventService port.DeviceEventProcessorInterface,
	log *zap.Logger,
) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			log.Info("Stopping background worker...")
			done := make(chan struct{})
			go func() {
				deviceEventService.StopProcessingGracefully()
				close(done)
			}()

			select {
			case <-done:
				log.Info("Background workers ha stopped")

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
			grpcServe(logger, config, grpcServer)

			checkCriticalDependencies(ctx, healthServer, logger, pool)

			return nil
		},
	})
}
