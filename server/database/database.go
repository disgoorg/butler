package database

import (
	"context"
	"embed"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/topi314/gomigrate"
	"github.com/topi314/gomigrate/drivers/postgres"

	"github.com/disgoorg/disgo-butler/server/database/sqlc"
)

//go:generate go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate

//go:embed migrations/*.sql
var migrations embed.FS

type Database struct {
	*sqlc.Queries
	pool *pgxpool.Pool
}

func New(ctx context.Context, cfg Config) (*Database, error) {
	pool, err := pgxpool.New(ctx, cfg.DataSourceName())
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err = pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	sqlDB := stdlib.OpenDBFromPool(pool)
	if err = gomigrate.Migrate(ctx, sqlDB, postgres.New, migrations, gomigrate.WithDirectory("migrations")); err != nil {
		_ = sqlDB.Close()
		pool.Close()
		return nil, fmt.Errorf("migrate database: %w", err)
	}
	if err = sqlDB.Close(); err != nil {
		pool.Close()
		return nil, fmt.Errorf("close migration connection: %w", err)
	}

	return &Database{
		Queries: sqlc.New(pool),
		pool:    pool,
	}, nil
}

func (d *Database) Close() {
	d.pool.Close()
}
