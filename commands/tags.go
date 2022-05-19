package commands

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo-butler/common"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/utils/paginator"
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
						Name:         "name",
						Description:  "the name of the tag to delete",
						Required:     true,
						Autocomplete: true,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "edit",
				Description: "let's you edit a tag",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionString{
						Name:         "name",
						Description:  "the name of the tag to edit",
						Required:     true,
						Autocomplete: true,
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
		"edit":   autoCompleteListTagHandler,
		"delete": autoCompleteListTagHandler,
		"list":   autoCompleteListTagHandler,
		"info":   autoCompleteListTagHandler,
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
	name := formatTagName(data.String("name"))
	tag, err := b.DB.Get(*e.GuildID(), name)
	if err == sql.ErrNoRows {
		return common.Respondf(e, "Tag `%s` does not exist.", name)
	} else if err != nil {
		return common.RespondMessageErr(e, "Failed to get tag info: ", err)
	}
	return e.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{
			{
				Title:       fmt.Sprintf("Tag `%s`", name),
				Description: tag.Content,
				Fields: []discord.EmbedField{
					{
						Name:  "Created by",
						Value: discord.UserMention(tag.OwnerID),
					},
					{
						Name:  "Uses",
						Value: strconv.Itoa(tag.Uses),
					},
					{
						Name:  "Created at",
						Value: fmt.Sprintf("%s (%s)", discord.NewTimestamp(discord.TimestampStyleNone, tag.CreatedAt), discord.NewTimestamp(discord.TimestampStyleRelative, tag.CreatedAt)),
					},
				},
			},
		},
	})
}

func listTagHandler(b *butler.Butler, e *events.ApplicationCommandInteractionEvent) error {
	tags, err := b.DB.GetAll(*e.GuildID())
	if err != nil {
		return common.RespondMessageErr(e, "Failed to list tags: ", err)
	}
	if len(tags) == 0 {
		return common.Respond(e, "No tags found.")
	}

	var pages []string
	curPage := ""
	for _, tag := range tags {
		newPage := fmt.Sprintf("**%s** - %s\n", tag.Name, discord.UserMention(tag.OwnerID))
		if len(curPage)+len(newPage) > 2000 {
			pages = append(pages, curPage)
			curPage = ""
		}
		curPage += newPage
	}
	if len(curPage) > 0 {
		pages = append(pages, curPage)
	}

	return b.Paginator.Create(e.Respond, &paginator.Paginator{
		PageFunc: func(page int, embed *discord.EmbedBuilder) {
			embed.SetDescription(pages[page])
		},
		MaxPages:        len(pages),
		ExpiryLastUsage: true,
		ID:              e.ID().String(),
	})
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
