package main

import (
	"context"

	"github.com/z0ne-dev/mgx/v2"
)

var Migrations = []mgx.Migration{
	mgx.NewMigration("1_create_device_events_table",
		func(ctx context.Context, tx mgx.Commands) error {
			_, err := tx.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS device_events (
					id UUID PRIMARY KEY,
					device_id VARCHAR(255) NOT NULL,
					device_name VARCHAR(255) NOT NULL,
					sensor_name VARCHAR(255) NOT NULL,
					message VARCHAR(255) NOT NULL,
					created_at TIMESTAMP(0) NOT NULL
				);
			`)

			return err
		},
	),
	mgx.NewMigration("2_create_idempotent_ids_table",
		func(ctx context.Context, tx mgx.Commands) error {
			_, err := tx.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS idempotent_ids (
					id UUID PRIMARY KEY,
					created_at TIMESTAMP(0) NOT NULL
				);
			`)

			return err
		},
	),
}
