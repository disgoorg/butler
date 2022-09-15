package commands

import (
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo-butler/common"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/handler"
)

func InfoCommand(b *butler.Butler) handler.Command {
	return handler.Command{
		Create: discord.SlashCommandCreate{
			Name:        "info",
			Description: "Provides information about disgo",
		},
		CommandHandlers: map[string]handler.CommandHandler{
			"": handleInfo(b),
		},
	}
}

func handleInfo(b *butler.Butler) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
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
}
