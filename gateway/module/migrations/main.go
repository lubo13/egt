package main

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	mgx "github.com/z0ne-dev/mgx/v2"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"git.infra.egt.com/gateway/module/infrastructure"
	"git.infra.egt.com/gateway/module/infrastructure/database"
)

func main() {
	fx.New(
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
		fx.Provide(
			func() context.Context { return context.Background() },
			zap.NewExample,
			infrastructure.NewConfig,
			database.NewDatabasePool,
		),
		fx.Invoke(func(log *zap.Logger) error {
			err := godotenv.Load()
			if err != nil {
				log.Fatal("Error loading .env file")
			}

			return err
		}),
		fx.Invoke(
			func(lc fx.Lifecycle, pool *pgxpool.Pool) {
				infrastructure.RegisterDatabaseLifecycleHooks(lc, pool)
			},
		),
		fx.Invoke(func(lifecycle fx.Lifecycle, shutdowner fx.Shutdowner, log *zap.Logger, pool *pgxpool.Pool) {
			lifecycle.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					go func() {
						exitCode := 0
						if err := execute(ctx, log, pool); err != nil {
							exitCode = 1
						}

						_ = shutdowner.Shutdown(fx.ExitCode(exitCode))
					}()
					return nil
				},
			})
		}),
	).Run()
}

func execute(ctx context.Context, log *zap.Logger, pool *pgxpool.Pool) error {
	migrator, err := mgx.New(
		mgx.Migrations(Migrations...),
	)
	if err != nil {
		log.Error("Unable to create migrator", zap.Error(err))

		return err
	}

	if err := migrator.Migrate(ctx, pool); err != nil {
		log.Error("Migration failed", zap.Error(err))

		return err
	}

	log.Info("Migrations applied successfully!")

	return nil
}
