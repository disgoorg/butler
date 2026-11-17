package server

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

const (
	colorError   = 0xFF0000
	colorSuccess = 0x5c5fea
)

func respondErr(respondFunc events.InteractionResponderFunc, err error) error {
	return respondErrMessage(respondFunc, fmt.Sprintf("Error while executing: %s", err))
}

func respondErrMessage(respondFunc events.InteractionResponderFunc, message string) error {
	return respondFunc(discord.InteractionResponseTypeCreateMessage, discord.NewMessageCreate().
		WithEmbeds(discord.NewEmbed().WithDescription(message).WithColor(colorError)).
		WithEphemeral(true),
	)
}

func respondErrMessagef(respondFunc events.InteractionResponderFunc, message string, a ...any) error {
	return respondErrMessage(respondFunc, fmt.Sprintf(message, a...))
}

func respondMessageErr(respondFunc events.InteractionResponderFunc, message string, err error) error {
	return respondFunc(discord.InteractionResponseTypeCreateMessage, discord.NewMessageCreate().
		WithEmbeds(discord.NewEmbed().WithDescriptionf(message, err).WithColor(colorError)).
		WithEphemeral(true),
	)
}

func respond(respondFunc events.InteractionResponderFunc, message string) error {
	return respondFunc(discord.InteractionResponseTypeCreateMessage, discord.NewMessageCreate().
		WithEmbeds(discord.NewEmbed().WithDescription(message).WithColor(colorSuccess)),
	)
}

func respondf(respondFunc events.InteractionResponderFunc, message string, a ...any) error {
	return respond(respondFunc, fmt.Sprintf(message, a...))
}
