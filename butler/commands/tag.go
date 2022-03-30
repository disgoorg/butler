package commands

import (
	"context"
	"database/sql"

	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo-butler/common"
	"github.com/disgoorg/disgo-butler/models"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func tagHandler(b *butler.Butler, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	var tag models.Tag

	_, err := b.DB.NewUpdate().
		Model((*models.Tag)(nil)).
		Set("uses = uses+?", 1).
		Where("guild_id = ?", *e.GuildID()).
		Where("name = ?", data.String("name")).
		Returning("content").
		Exec(context.TODO(), &tag)
	if err == sql.ErrNoRows {
		return common.Respond(e, "Tag not found")
	} else if err != nil {
		return common.RespondErr(e, err)
	}
	return e.CreateMessage(discord.MessageCreate{
		Content: tag.Content,
	})
}
