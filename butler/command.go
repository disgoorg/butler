package butler

import (
	"github.com/DisgoOrg/disgo/core/events"
	"github.com/DisgoOrg/disgo/discord"
)

type Command struct {
	Definition          discord.ApplicationCommandCreate
	Handler             func(b *Butler, e *events.ApplicationCommandInteractionEvent)
	AutocompleteHandler func(b *Butler, e *events.AutocompleteInteractionEvent)
}
