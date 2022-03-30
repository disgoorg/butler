package commands

import (
	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func handlePing(_ *butler.Butler, e *events.ApplicationCommandInteractionEvent) error {
	return e.CreateMessage(discord.MessageCreate{Content: "Pong"})
}
