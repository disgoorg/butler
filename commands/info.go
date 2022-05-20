package commands

import (
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo-butler/common"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

var InfoCommand = butler.Command{
	Create: discord.SlashCommandCreate{
		CommandName: "info",
		Description: "Provides information about disgo",
	},
	CommandHandlers: map[string]butler.HandleFunc{
		"": handleInfo,
	},
}

func handleInfo(b *butler.Butler, e *events.ApplicationCommandInteractionEvent) error {
	user, _ := b.Client.Caches().GetSelfUser()
	return e.CreateMessage(discord.NewMessageCreateBuilder().
		SetEmbeds(discord.NewEmbedBuilder().
			SetTitle("DisGo Butler").
			SetThumbnail(user.EffectiveAvatarURL()).
			SetColor(common.ColorSuccess).
			AddField("Version", b.Version, false).
			AddField("DisGo Version", disgo.Version, false).
			Build(),
		).
		AddActionRow(discord.NewLinkButton("GitHub", "https://github.com/disgoorg/disgo-butler")).
		Build(),
	)
}
