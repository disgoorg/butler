package commands

import (
	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func handleInfo(b *butler.Butler, e *events.ApplicationCommandInteractionEvent) error {
	return e.CreateMessage(discord.NewMessageCreateBuilder().
		SetEmbeds(discord.NewEmbedBuilder().
			AddField("Version", "", false).
			Build(),
		).
		Build(),
	)
}
