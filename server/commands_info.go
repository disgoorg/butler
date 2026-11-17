package server

import (
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

var infoCommand = discord.SlashCommandCreate{
	Name:        "info",
	Description: "Provides information about disgo",
}

func (s *Server) handleInfo(event *handler.CommandEvent) error {
	user, _ := s.Client.Caches.SelfUser()
	return event.CreateMessage(discord.NewMessageCreate().
		WithEmbeds(discord.NewEmbed().
			WithTitle("DisGo Butler").
			WithThumbnail(user.EffectiveAvatarURL()).
			WithColor(colorSuccess).
			AddField("Version", s.Version, false).
			AddField("DisGo Version", disgo.Version, false),
		).
		AddActionRow(discord.NewLinkButton("GitHub", "https://github.com/disgoorg/disgo-butler")),
	)
}
