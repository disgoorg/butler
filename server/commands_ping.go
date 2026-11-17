package server

import (
	"log/slog"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

var pingCommand = discord.SlashCommandCreate{
	Name:        "ping",
	Description: "Responds with pong",
}

func (s *Server) handlePing(event *handler.CommandEvent) error {
	var gatewayPing string
	if event.Client().HasGateway() {
		gatewayPing = event.Client().Gateway.Latency().String()
	}

	embed := discord.NewEmbed().
		WithTitle("Ping").
		AddField("Rest", "loading...", false).
		AddField("Gateway", gatewayPing, false).
		WithColor(colorSuccess)

	defer func() {
		start := time.Now()
		_, _ = event.Client().Rest.GetBotApplicationInfo()
		embed = embed.WithField(0, "Rest", time.Since(start).String(), false)
		if _, err := event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.NewMessageUpdate().WithEmbeds(embed)); err != nil {
			event.Client().Logger.Error("failed to update ping embed", slog.Any("err", err))
		}
	}()

	return event.CreateMessage(discord.NewMessageCreate().WithEmbeds(embed))
}
