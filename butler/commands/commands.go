package commands

import (
	"github.com/DisgoOrg/disgo-butler/butler"
	"github.com/DisgoOrg/disgo/discord"
)

var Commands = []butler.Command{
	{
		Create: discord.SlashCommandCreate{
			Name:              "ping",
			Description:       "Responds with pong",
			DefaultPermission: true,
		},
		CommandHandlers: map[string]butler.HandleFunc{
			"": handlePing,
		},
	},
	{
		Create: discord.SlashCommandCreate{
			Name:              "info",
			Description:       "Provides information about disgo",
			DefaultPermission: true,
		},
		CommandHandlers: map[string]butler.HandleFunc{
			"": handleInfo,
		},
	},
	{
		Create: discord.SlashCommandCreate{
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
		CommandHandlers: map[string]butler.HandleFunc{
			"": handleDocs,
		},
		AutocompleteHandlers: map[string]butler.AutocompleteHandleFunc{
			"": handleDocsAutocomplete,
		},
	},
	{
		Create: discord.SlashCommandCreate{
			Name:        "tag",
			Description: "let's you display a tag",
			Options: []discord.ApplicationCommandOption{

				discord.ApplicationCommandOptionString{
					Name:        "name",
					Description: "the name of the tag to display",
					Required:    true,
				},
			},
			DefaultPermission: true,
		},
		CommandHandlers: map[string]butler.HandleFunc{
			"": tagHandler,
		},
	},
	{
		Create: discord.SlashCommandCreate{
			Name:        "tags",
			Description: "let's you create/delete/edit tags",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionSubCommand{
					Name:        "create",
					Description: "let's you create a tag",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:        "name",
							Description: "the name of the tag to create",
							Required:    true,
						},
						discord.ApplicationCommandOptionString{
							Name:        "content",
							Description: "the content of the new tag",
							Required:    true,
						},
					},
				},
				discord.ApplicationCommandOptionSubCommand{
					Name:        "delete",
					Description: "let's you delete a tag",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:        "name",
							Description: "the name of the tag to delete",
							Required:    true,
						},
					},
				},
				discord.ApplicationCommandOptionSubCommand{
					Name:        "edit",
					Description: "let's you edit a tag",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:        "name",
							Description: "the name of the tag to edit",
							Required:    true,
						},
						discord.ApplicationCommandOptionString{
							Name:        "content",
							Description: "the new content of the new tag",
							Required:    true,
						},
					},
				},
			},
			DefaultPermission: true,
		},
		CommandHandlers: map[string]butler.HandleFunc{
			"create": createTagHandler,
			"delete": deleteTagHandler,
			"edit":   editTagHandler,
		},
	},
}
