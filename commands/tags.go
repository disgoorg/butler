package commands

import (
	"strings"

	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo-butler/common"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

var TagsCommand = butler.Command{
	Create: discord.SlashCommandCreate{
		CommandName: "tags",
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
			discord.ApplicationCommandOptionSubCommand{
				Name:        "info",
				Description: "lets you view a tag's info",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionString{
						Name:         "name",
						Description:  "the name of the tag to view",
						Required:     true,
						Autocomplete: true,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "list",
				Description: "lists all tags",
			},
		},
	},
	CommandHandlers: map[string]butler.HandleFunc{
		"create": createTagHandler,
		"edit":   editTagHandler,
		"delete": deleteTagHandler,
		"info":   infoTagHandler,
		"list":   listTagHandler,
	},
	AutocompleteHandlers: map[string]butler.AutocompleteHandleFunc{
		"list": autoCompleteListTagHandler,
		"info": autoCompleteListTagHandler,
	},
}

func createTagHandler(b *butler.Butler, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	if err := b.DB.Create(*e.GuildID(), e.User().ID, formatTagName(data.String("name")), data.String("content")); err != nil {
		return common.RespondMessageErr(e, "Failed to create tag: ", err)
	}
	return common.Respond(e, "Tag created!")
}

func editTagHandler(b *butler.Butler, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	if err := b.DB.Edit(*e.GuildID(), formatTagName(data.String("name")), data.String("content")); err != nil {
		return common.RespondMessageErr(e, "Failed to edit tag: ", err)
	}
	return common.Respond(e, "Tag edited.")
}

func deleteTagHandler(b *butler.Butler, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	if err := b.DB.Delete(*e.GuildID(), formatTagName(data.String("name"))); err != nil {
		return common.RespondMessageErr(e, "Failed to delete tag: ", err)
	}
	return common.Respond(e, "Tag deleted.")
}

func infoTagHandler(b *butler.Butler, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	if err := b.DB.Delete(*e.GuildID(), formatTagName(data.String("name"))); err != nil {
		return common.RespondMessageErr(e, "Failed to delete tag: ", err)
	}
	return common.Respond(e, "Tag deleted.")
}

func listTagHandler(b *butler.Butler, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	if err := b.DB.Delete(*e.GuildID(), formatTagName(data.String("name"))); err != nil {
		return common.RespondMessageErr(e, "Failed to delete tag: ", err)
	}
	return common.Respond(e, "Tag deleted.")
}

func autoCompleteListTagHandler(b *butler.Butler, e *events.AutocompleteInteractionEvent) error {
	name := formatTagName(e.Data.String("name"))

	tags, err := b.DB.GetAll(*e.GuildID())
	if err != nil {
		return e.Result(nil)
	}
	var response []discord.AutocompleteChoice

	options := make([]string, len(tags))
	for i := range tags {
		options[i] = tags[i].Name
	}
	options = fuzzy.FindFold(name, options)
	for _, option := range options {
		if len(response) >= 25 {
			break
		}
		response = append(response, discord.AutocompleteChoiceString{
			Name:  option,
			Value: option,
		})
	}
	return e.Result(response)
}

func formatTagName(name string) string {
	return strings.ToLower(name)
}
