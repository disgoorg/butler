package common

import (
	"github.com/DisgoOrg/disgo/core/events"
	"github.com/DisgoOrg/disgo/discord"
)

const (
	ColorError   = 0xFF0000
	ColorSuccess = 0x5c5fea
)

func RespondErr(e *events.ApplicationCommandInteractionEvent, err error) error {
	return e.Create(discord.NewMessageCreateBuilder().
		SetEmbeds(discord.NewEmbedBuilder().
			SetDescriptionf("Error while executing: %s", err).
			SetColor(ColorError).
			Build(),
		).
		Build(),
	)
}

func Respond(e *events.ApplicationCommandInteractionEvent, message string) error {
	return e.Create(discord.NewMessageCreateBuilder().
		SetEmbeds(discord.NewEmbedBuilder().
			SetDescriptionf(message).
			SetColor(ColorSuccess).
			Build(),
		).
		Build(),
	)
}
