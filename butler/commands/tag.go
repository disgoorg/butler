package commands

import (
	"context"

	"github.com/DisgoOrg/disgo-butler/butler"
	"github.com/DisgoOrg/disgo-butler/common"
	"github.com/DisgoOrg/disgo-butler/models"
	"github.com/DisgoOrg/disgo/core/events"
	"github.com/DisgoOrg/disgo/discord"
)

func tagHandler(b *butler.Butler, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	var tag models.Tag

	_, err := b.DB.NewUpdate().
		Model((*models.Tag)(nil)).
		Set("uses = uses+?", 1).
		Where("guild_id = ?", *e.GuildID).
		Where("name = ?", *data.Options.String("name")).
		Returning("content").
		Exec(context.TODO(), &tag)

	if err != nil {
		return common.RespondErr(e, err)
	}
	return e.Create(discord.MessageCreate{
		Content: tag.Content,
	})
}
