package db

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

type Config struct {
	Address  string `json:"address"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	Insecure bool   `json:"insecure"`
	Verbose  bool   `json:"verbose"`
}

func SetupDatabase(config Config) (DB, error) {
	db := bun.NewDB(sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithAddr(config.Address),
		pgdriver.WithUser(config.User),
		pgdriver.WithPassword(config.Password),
		pgdriver.WithDatabase(config.Database),
		pgdriver.WithInsecure(config.Insecure),
	)), pgdialect.New())

	db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(config.Verbose)))

	if _, err := db.NewCreateTable().Model((*Tag)(nil)).Exec(context.TODO()); err != nil {
		return nil, err
	}

	return &sqlDB{db: db}, nil
}

type DB interface {
	TagsDB
}

type sqlDB struct {
	db *bun.DB
}
