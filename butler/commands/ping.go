package commands

import (
	"github.com/DisgoOrg/disgo-butler/butler"
	"github.com/DisgoOrg/disgo/core/events"
	"github.com/DisgoOrg/disgo/discord"
)

func handlePing(_ *butler.Butler, e *events.ApplicationCommandInteractionEvent) error {
	return e.Create(discord.MessageCreate{Content: "Pong"})
}
