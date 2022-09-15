package commands

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo-butler/common"
	"github.com/disgoorg/disgo-butler/db"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/handler"
	"github.com/disgoorg/utils/paginator"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

func TagsCommand(b *butler.Butler) handler.Command {
	return handler.Command{
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
		CommandHandlers: map[string]handler.CommandHandler{
			"create": createTagHandler(b),
			"edit":   editTagHandler(b),
			"delete": deleteTagHandler(b),
			"info":   infoTagHandler(b),
			"list":   listTagHandler(b),
		},
		AutocompleteHandlers: map[string]handler.AutocompleteHandler{
			"edit":   autoCompleteListTagHandler(b, true),
			"delete": autoCompleteListTagHandler(b, true),
			"list":   autoCompleteListTagHandler(b, false),
			"info":   autoCompleteListTagHandler(b, false),
		},
	}
}

func createTagHandler(b *butler.Butler) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		data := e.SlashCommandInteractionData()
		name := formatTagName(data.String("name"))

		if _, err := b.DB.Get(*e.GuildID(), name); err == nil {
			return common.RespondErrMessage(e.Respond, "Tag already exists.")
		} else if err != nil && err != sql.ErrNoRows {
			return common.RespondMessageErr(e.Respond, "Failed to edit tag: %s", err)
		}

		if err := b.DB.Create(*e.GuildID(), e.User().ID, name, data.String("content")); err != nil {
			return common.RespondMessageErr(e.Respond, "Failed to create tag: %s", err)
		}
		return common.Respond(e.Respond, "Tag created!")
	}
}

func editTagHandler(b *butler.Butler) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		data := e.SlashCommandInteractionData()
		name := formatTagName(data.String("name"))

		tag, err := b.DB.Get(*e.GuildID(), name)
		if err == sql.ErrNoRows {
			return common.RespondErrMessage(e.Respond, "Tag not found.")
		} else if err != nil {
			return common.RespondMessageErr(e.Respond, "Failed to edit tag: %s", err)
		}
		if e.User().ID != tag.OwnerID && e.Member().Permissions.Missing(discord.PermissionManageServer) {
			return common.RespondErrMessage(e.Respond, "You do not have permission to edit this tag.")
		}

		if err = b.DB.Edit(*e.GuildID(), name, data.String("content")); err != nil {
			return common.RespondMessageErr(e.Respond, "Failed to edit tag: %s", err)
		}
		return common.Respond(e.Respond, "Tag edited.")
	}
}

func deleteTagHandler(b *butler.Butler) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		data := e.SlashCommandInteractionData()
		name := formatTagName(data.String("name"))

		tag, err := b.DB.Get(*e.GuildID(), name)
		if err == sql.ErrNoRows {
			return common.RespondErrMessage(e.Respond, "Tag not found.")
		} else if err != nil {
			return common.RespondMessageErr(e.Respond, "Failed to delete tag: %s", err)
		}
		if e.User().ID != tag.OwnerID && e.Member().Permissions.Missing(discord.PermissionManageServer) {
			return common.RespondErrMessage(e.Respond, "You do not have permission to delete this tag.")
		}

		if err = b.DB.Delete(*e.GuildID(), name); err != nil {
			return common.RespondMessageErr(e.Respond, "Failed to delete tag: %s", err)
		}
		return common.Respond(e.Respond, "Tag deleted.")
	}
}

func infoTagHandler(b *butler.Butler) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		data := e.SlashCommandInteractionData()
		name := formatTagName(data.String("name"))
		tag, err := b.DB.Get(*e.GuildID(), name)
		if err == sql.ErrNoRows {
			return common.Respondf(e.Respond, "Tag `%s` does not exist.", name)
		} else if err != nil {
			return common.RespondMessageErr(e.Respond, "Failed to get tag info: ", err)
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
}

func listTagHandler(b *butler.Butler) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		tags, err := b.DB.GetAll(*e.GuildID())
		if err != nil {
			return common.RespondMessageErr(e.Respond, "Failed to list tags: ", err)
		}
		if len(tags) == 0 {
			return common.Respond(e.Respond, "No tags found.")
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
}

func autoCompleteListTagHandler(b *butler.Butler, filterTags bool) handler.AutocompleteHandler {
	return func(e *events.AutocompleteInteractionCreate) error {
		name := formatTagName(e.Data.String("name"))

		var (
			tags []db.Tag
			err  error
		)
		if filterTags && e.Member().Permissions.Missing(discord.PermissionManageServer) {
			tags, err = b.DB.GetAllUser(*e.GuildID(), e.User().ID)
		} else {
			tags, err = b.DB.GetAll(*e.GuildID())
		}

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
}

func formatTagName(name string) string {
	return strings.ToLower(name)
}
