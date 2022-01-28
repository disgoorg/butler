package commands

import (
	"github.com/DisgoOrg/disgo-butler/butler"
	"github.com/DisgoOrg/disgo/discord"
)

var Commands = []butler.Command{
	{
		Definition: discord.SlashCommandCreate{
			Name:              "ping",
			Description:       "Responds with pong",
			DefaultPermission: true,
		},
		Handler: handlePing,
	},
	{
		Definition: discord.SlashCommandCreate{
			Name:              "info",
			Description:       "Provides information about disgo",
			DefaultPermission: true,
		},
		Handler: handleInfo,
	},
	{
		Definition: discord.SlashCommandCreate{
			Name:              "docs",
			Description:       "Provides info to the provided module, type, function, etc.",
			DefaultPermission: true,
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:         "module",
					Description:  "The module to lookup. Example: github.com/DisgoOrg/disgo/core",
					Required:     true,
					Autocomplete: true,
				},
				discord.ApplicationCommandOptionString{
					Name:         "query",
					Description:  "The lookup query. Example: MessageCreate",
					Autocomplete: true,
				},
			},
		},
		Handler:             handleDocs,
		AutocompleteHandler: handleDocsAutocomplete,
	},
}
