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

func RespondErrMessagef(e *events.ApplicationCommandInteractionEvent, message string, a ...any) error {
	return RespondErrMessage(e, fmt.Sprintf(message, a...))
}

func RespondMessageErr(e *events.ApplicationCommandInteractionEvent, message string, err error) error {
	return e.CreateMessage(discord.NewMessageCreateBuilder().
		SetEmbeds(discord.NewEmbedBuilder().
			SetDescriptionf(message, err).
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

func Respondf(e *events.ApplicationCommandInteractionEvent, message string, a ...any) error {
	return Respond(e, fmt.Sprintf(message, a...))
}
