package butler

import (
	"context"
	"database/sql"

	"github.com/disgoorg/disgo-butler/models"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

func (b *Butler) SetupDatabase() error {
	sqlDB := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithAddr(b.Config.Database.Address),
		pgdriver.WithUser(b.Config.Database.User),
		pgdriver.WithPassword(b.Config.Database.Password),
		pgdriver.WithDatabase(b.Config.Database.Database),
		pgdriver.WithInsecure(b.Config.Database.Insecure),
	))
	b.DB = bun.NewDB(sqlDB, pgdialect.New())
	b.DB.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(b.Config.Database.Verbose),
	))

	if _, err := b.DB.NewCreateTable().Model((*models.Tag)(nil)).Exec(context.TODO()); err != nil {
		return err
	}

	return nil
}
