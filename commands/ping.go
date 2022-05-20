package commands

import (
	"time"

	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo-butler/common"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/rest"
)

var PingCommand = butler.Command{
	Create: discord.SlashCommandCreate{
		CommandName: "ping",
		Description: "Responds with pong",
	},
	CommandHandlers: map[string]butler.HandleFunc{
		"": handlePing,
	},
}

func handlePing(_ *butler.Butler, e *events.ApplicationCommandInteractionEvent) error {
	var gatewayPing string
	if e.Client().HasGateway() {
		gatewayPing = e.Client().Gateway().Latency().String()
	}

	eb := discord.NewEmbedBuilder().
		SetTitle("Ping").
		AddField("Rest", "loading...", false).
		AddField("Gateway", gatewayPing, false).
		SetColor(common.ColorSuccess)

	defer func() {
		var start int64
		_, _ = e.Client().Rest().GetBotApplicationInfo(func(config *rest.RequestConfig) {
			start = time.Now().UnixNano()
		})
		duration := time.Now().UnixNano() - start
		eb.SetField(0, "Rest", time.Duration(duration).String(), false)
		if _, err := e.Client().Rest().UpdateInteractionResponse(e.ApplicationID(), e.Token(), discord.MessageUpdate{Embeds: &[]discord.Embed{eb.Build()}}); err != nil {
			e.Client().Logger().Error("Failed to update ping embed: ", err)
		}
	}()

	return e.Respond(discord.InteractionCallbackTypeCreateMessage, discord.NewMessageCreateBuilder().
		SetEmbeds(eb.Build()).
		Build(),
	)
}
