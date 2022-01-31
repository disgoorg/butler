package models

import (
	"time"

	"github.com/DisgoOrg/snowflake"
)

type Tag struct {
	ID        int                 `bun:"id,type:bigserial,pk,notnull"`
	GuildID   snowflake.Snowflake `bun:"guild_id,notnull,unique:name-guild"`
	Name      string              `bun:"name,notnull,unique:name-guild"`
	Content   string              `bun:"content,notnull"`
	Owner     snowflake.Snowflake `bun:"owner,notnull"`
	Uses      int                 `bun:"uses,notnull"`
	CreatedAt time.Time           `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt time.Time           `bun:"updated_at,notnull,default:current_timestamp"`
}
