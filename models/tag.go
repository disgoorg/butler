package models

import (
	"time"

	"github.com/disgoorg/snowflake"
)

type Tag struct {
	GuildID   snowflake.Snowflake `bun:"guild_id,pk"`
	Name      string              `bun:"name,pk"`
	Content   string              `bun:"content,notnull"`
	Owner     snowflake.Snowflake `bun:"owner,notnull"`
	Uses      int                 `bun:"uses,notnull"`
	CreatedAt time.Time           `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt time.Time           `bun:"updated_at,notnull,default:current_timestamp"`
}
