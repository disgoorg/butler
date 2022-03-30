package common

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

const (
	ColorError   = 0xFF0000
	ColorSuccess = 0x5c5fea
)

func RespondErr(e *events.ApplicationCommandInteractionEvent, err error) error {
	return RespondErrMessage(e, fmt.Sprintf("Error while executing: %s", err))
}

func RespondErrMessage(e *events.ApplicationCommandInteractionEvent, message string) error {
	return e.CreateMessage(discord.NewMessageCreateBuilder().
		SetEmbeds(discord.NewEmbedBuilder().
			SetDescription(message).
			SetColor(ColorError).
			Build(),
		).
		SetEphemeral(true).
		Build(),
	)
}

func Respond(e *events.ApplicationCommandInteractionEvent, message string) error {
	return e.CreateMessage(discord.NewMessageCreateBuilder().
		SetEmbeds(discord.NewEmbedBuilder().
			SetDescription(message).
			SetColor(ColorSuccess).
			Build(),
		).
		Build(),
	)
}
