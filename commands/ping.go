package commands

import (
	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
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
	return e.CreateMessage(discord.MessageCreate{Content: "Pong"})
}
