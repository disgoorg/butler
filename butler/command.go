package butler

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func (b *Butler) SetupCommands(shouldSyncCommands bool, commands ...Command) {
	commandCreates := make([]discord.ApplicationCommandCreate, len(commands))
	for i, command := range commands {
		commandCreates[i] = command.Create
		b.Commands[command.Create.Name()] = command
	}

	if shouldSyncCommands {
		b.Client.Logger().Info("Syncing commands...")
		var err error
		if b.Config.DevMode {
			_, err = b.Client.Rest().SetGuildCommands(b.Client.ApplicationID(), b.Config.GuildID, commandCreates)
		} else {
			_, err = b.Client.Rest().SetGlobalCommands(b.Client.ApplicationID(), commandCreates)
		}
		if err != nil {
			b.Client.Logger().Error("Failed to set commands: ", err)
		}
	}
}

func (b *Butler) OnApplicationCommandInteraction(e *events.ApplicationCommandInteractionEvent) {
	if command, ok := b.Commands[e.Data.CommandName()]; ok {
		var path string
		if data, ok := e.Data.(discord.SlashCommandInteractionData); ok {
			if data.SubCommandGroupName != nil {
				path += *data.SubCommandGroupName + "/"
			}
			if data.SubCommandName != nil {
				path += *data.SubCommandName
			}
		}
		if handler, ok := command.CommandHandlers[path]; ok {
			if err := handler(b, e); err != nil {
				b.Client.Logger().Error("Error handling command: ", err)
			}
			return
		}
		b.Logger.Warnf("No handler for command with path %s found", path)
		return
	}
	b.Logger.Warnf("No handler for command with name %s found", e.Data.CommandName())
}

func (b *Butler) OnAutocompleteInteraction(e *events.AutocompleteInteractionEvent) {
	if command, ok := b.Commands[e.Data.CommandName]; ok {
		var path string
		if e.Data.SubCommandGroupName != nil {
			path += *e.Data.SubCommandGroupName + "/"
		}
		if e.Data.SubCommandName != nil {
			path += *e.Data.SubCommandName
		}

		if handler, ok := command.AutocompleteHandlers[path]; ok {
			if err := handler(b, e); err != nil {
				b.Client.Logger().Error("Error handling autocomplete: ", err)
			}
			return
		}
		b.Logger.Warnf("No handler for autocomplete with path %s found", path)
		return
	}
	b.Logger.Warnf("No handler for autocomplete with name %s found", e.Data.CommandName)
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
