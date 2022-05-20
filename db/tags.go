package db

import (
	"context"
	"time"

	"github.com/disgoorg/snowflake/v2"
)

type TagsDB interface {
	Get(guildID snowflake.ID, name string) (Tag, error)
	GetAndIncrement(guildID snowflake.ID, name string) (Tag, error)
	GetAll(guildID snowflake.ID) ([]Tag, error)
	GetAllUser(guildID snowflake.ID, userID snowflake.ID) ([]Tag, error)
	Create(guildID snowflake.ID, ownerID snowflake.ID, name string, content string) error
	Edit(guildID snowflake.ID, name string, content string) error
	Delete(guildID snowflake.ID, name string) error
}

type Tag struct {
	GuildID   snowflake.ID `bun:"guild_id,pk"`
	Name      string       `bun:"name,pk"`
	Content   string       `bun:"content,notnull"`
	OwnerID   snowflake.ID `bun:"owner_id,notnull"`
	Uses      int          `bun:"uses,notnull"`
	CreatedAt time.Time    `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt time.Time    `bun:"updated_at,notnull,default:current_timestamp"`
}

func (s *sqlDB) Get(guildID snowflake.ID, name string) (tag Tag, err error) {
	err = s.db.NewSelect().
		Model(&tag).
		Where("guild_id = ?", guildID).
		Where("name = ?", name).
		Scan(context.TODO())
	return
}

func (s *sqlDB) GetAndIncrement(guildID snowflake.ID, name string) (tag Tag, err error) {
	_, err = s.db.NewUpdate().
		Model(&tag).
		Set("uses = uses+?", 1).
		Where("guild_id = ?", guildID).
		Where("name = ?", name).
		Returning("content").
		Exec(context.TODO())
	return
}

func (s *sqlDB) GetAll(guildID snowflake.ID) (tags []Tag, err error) {
	err = s.db.NewSelect().
		Model(&tags).
		Where("guild_id = ?", guildID).
		Scan(context.TODO())
	return
}

func (s *sqlDB) GetAllUser(guildID snowflake.ID, userID snowflake.ID) (tags []Tag, err error) {
	err = s.db.NewSelect().
		Model(&tags).
		Where("guild_id = ? AND owner_id = ?", guildID, userID).
		Scan(context.TODO())
	return
}

func (s *sqlDB) Create(guildID snowflake.ID, ownerID snowflake.ID, name string, content string) (err error) {
	_, err = s.db.NewInsert().Model(&Tag{
		GuildID: guildID,
		Name:    name,
		OwnerID: ownerID,
		Content: content,
	}).Exec(context.TODO())
	return
}

func (s *sqlDB) Edit(guildID snowflake.ID, name string, content string) (err error) {
	_, err = s.db.NewUpdate().Model((*Tag)(nil)).Set("content = ?", content).Where(" guild_id = ? AND name = ?", guildID, name).Exec(context.TODO())
	return
}

func (s *sqlDB) Delete(guildID snowflake.ID, name string) (err error) {
	_, err = s.db.NewDelete().Model((*Tag)(nil)).Where(" guild_id = ? AND name = ?", guildID, name).Exec(context.TODO())
	return
}
