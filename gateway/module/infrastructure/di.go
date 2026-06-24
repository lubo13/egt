package infrastructure

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

func RegisterDatabaseLifecycleHooks(lc fx.Lifecycle, pool *pgxpool.Pool) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Ping the database to ensure the connection is working
			return pool.Ping(ctx)
		},
		OnStop: func(ctx context.Context) error {
			// Close the pool when the application stops
			pool.Close()
			return nil
		},
	})
}
