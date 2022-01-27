package common

import (
	"github.com/DisgoOrg/disgo/core/events"
	"github.com/DisgoOrg/disgo/discord"
	"github.com/DisgoOrg/log"
)

func RespondError(e *events.ApplicationCommandInteractionEvent, errr error) {
	if err := e.Create(discord.NewMessageCreateBuilder().
		SetEmbeds(discord.NewEmbedBuilder().
			SetDescriptionf("Error while executing: %s", errr).
			Build(),
		).
		Build(),
	); err != nil {
		log.Error(err)
	}
}
