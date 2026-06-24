package server

import (
	"context"
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"git.infra.egt.com/producer/module/infrastructure"
)

type PingInterface interface {
	Ping(ctx context.Context) error
}

func NewGrpcServer(logger *zap.Logger) *grpc.Server {
	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)

	return grpcServer
}

func NewGrpcHealthServer(logger *zap.Logger, grpcServer *grpc.Server) *health.Server {
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	return healthServer
}

func GrpcServe(logger *zap.Logger, config *infrastructure.Config, grpcServer *grpc.Server) {
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%s", config.GRPCPort))
		if err != nil {
			logger.Error("Failed to listen", zap.Error(err))
		}

		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("Failed to serve", zap.Error(err))
		}

		logger.Info("GRPC health server is ready")
	}()
}

func CheckCriticalDependencies(
	ctx context.Context,
	healthServer *health.Server,
	logger *zap.Logger,
	pingCollection ...PingInterface,
) {
	go func(
		ctx context.Context,
		healthServer *health.Server,
		logger *zap.Logger,
		pingCollection ...PingInterface,
	) {
		logger.Info("grpc health check is running...")

		for {
			for _, pinger := range pingCollection {
				ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
				if err := pinger.Ping(ctx); err == nil {
					healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
				} else {
					healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)

					err = pinger.Ping(ctx)
					logger.Error("malfunction of critical dependency", zap.Error(err))
				}
			}

			time.Sleep(5 * time.Second)
		}
	}(ctx, healthServer, logger, pingCollection...)
}
