package butler

import (
	"github.com/DisgoOrg/disgo/core"
	"github.com/DisgoOrg/disgo/core/events"
	"github.com/DisgoOrg/disgo/discord"
	"github.com/DisgoOrg/log"
)

func (b *Butler) OnApplicationCommandInteraction(e *events.ApplicationCommandInteractionEvent) {
	if command, ok := b.Commands[e.Data.ID()]; ok {
		var path string
		if data, ok := e.Data.(*core.SlashCommandInteractionData); ok {
			if name := data.SubCommandName; name != nil {
				path += "/" + *name
			}
			if name := data.SubCommandGroupName; name != nil {
				path += "/" + *name
			}
		}
		if handler, ok := command.CommandHandlers[path]; ok {
			if err := handler(b, e); err != nil {
				b.Bot.Logger.Error("Error handling command: ", err)
			}
		}
		return
	}
	log.Warnf("No handler for command with ID %s found", e.Data.ID())
}

func (b *Butler) OnAutocompleteInteraction(e *events.AutocompleteInteractionEvent) {
	if command, ok := b.Commands[e.Data.CommandID]; ok {
		var path string
		if name := e.Data.SubCommandName; name != nil {
			path += "/" + *name
		}
		if name := e.Data.SubCommandGroupName; name != nil {
			path += "/" + *name
		}

		if handler, ok := command.AutocompleteHandlers[path]; ok {
			if err := handler(b, e); err != nil {
				b.Bot.Logger.Error("Error handling autocomplete: ", err)
			}
		}
		return
	}
	log.Warnf("No handler for autocomplete with ID %s found", e.Data.CommandID)
}

type (
	HandleFunc             func(b *Butler, e *events.ApplicationCommandInteractionEvent) error
	AutocompleteHandleFunc func(b *Butler, e *events.AutocompleteInteractionEvent) error
	Command                struct {
		Create               discord.ApplicationCommandCreate
		CommandHandlers      map[string]HandleFunc
		AutocompleteHandlers map[string]AutocompleteHandleFunc
	}
)
