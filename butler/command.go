package butler

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/log"
)

func (b *Butler) OnApplicationCommandInteraction(e *events.ApplicationCommandInteractionEvent) {
	if command, ok := b.Commands[e.Data.CommandID()]; ok {
		var path string
		if data, ok := e.Data.(discord.SlashCommandInteractionData); ok {
			if name := data.SubCommandName; name != nil {
				path += "/" + *name
			}
			if name := data.SubCommandGroupName; name != nil {
				path += "/" + *name
			}
		}
		if handler, ok := command.CommandHandlers[path]; ok {
			if err := handler(b, e); err != nil {
				b.Client.Logger().Error("Error handling command: ", err)
			}
		}
		return
	}
	log.Warnf("No handler for command with ID %s found", e.Data.CommandID())
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
				b.Client.Logger().Error("Error handling autocomplete: ", err)
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
