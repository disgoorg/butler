package commands

import (
	"github.com/DisgoOrg/disgo-butler/butler"
	"github.com/DisgoOrg/disgo/core/events"
	"github.com/DisgoOrg/disgo/discord"
)

func handleInfo(b *butler.Butler, e *events.ApplicationCommandInteractionEvent) {
	if err := e.Create(discord.NewMessageCreateBuilder().
		SetEmbeds(discord.NewEmbedBuilder().
			AddField("Version", "", false).
			Build(),
		).
		Build(),
	); err != nil {
		b.Bot.Logger.Error(err)
	}
}
