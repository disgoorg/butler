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

func createTagHandler(b *butler.Butler) func(ctx *handler.CommandContext) error {
	return func(ctx *handler.CommandContext) error {
		data := ctx.SlashCommandInteractionData()
		name := formatTagName(data.String("name"))

		if _, err := b.DB.Get(*ctx.GuildID(), name); err == nil {
			return common.RespondErrMessage(ctx.Respond, "Tag already exists.")
		} else if err != nil && err != sql.ErrNoRows {
			return common.RespondMessageErr(ctx.Respond, "Failed to edit tag: %s", err)
		}

		if err := b.DB.Create(*ctx.GuildID(), ctx.User().ID, name, data.String("content")); err != nil {
			return common.RespondMessageErr(ctx.Respond, "Failed to create tag: %s", err)
		}
		return common.Respond(ctx.Respond, "Tag created!")
	}
}

func editTagHandler(b *butler.Butler) func(ctx *handler.CommandContext) error {
	return func(ctx *handler.CommandContext) error {
		data := ctx.SlashCommandInteractionData()
		name := formatTagName(data.String("name"))

		tag, err := b.DB.Get(*ctx.GuildID(), name)
		if err == sql.ErrNoRows {
			return common.RespondErrMessage(ctx.Respond, "Tag not found.")
		} else if err != nil {
			return common.RespondMessageErr(ctx.Respond, "Failed to edit tag: %s", err)
		}
		if ctx.User().ID != tag.OwnerID && ctx.Member().Permissions.Missing(discord.PermissionManageServer) {
			return common.RespondErrMessage(ctx.Respond, "You do not have permission to edit this tag.")
		}

		if err = b.DB.Edit(*ctx.GuildID(), name, data.String("content")); err != nil {
			return common.RespondMessageErr(ctx.Respond, "Failed to edit tag: %s", err)
		}
		return common.Respond(ctx.Respond, "Tag edited.")
	}
}

func deleteTagHandler(b *butler.Butler) func(ctx *handler.CommandContext) error {
	return func(ctx *handler.CommandContext) error {
		data := ctx.SlashCommandInteractionData()
		name := formatTagName(data.String("name"))

		tag, err := b.DB.Get(*ctx.GuildID(), name)
		if err == sql.ErrNoRows {
			return common.RespondErrMessage(ctx.Respond, "Tag not found.")
		} else if err != nil {
			return common.RespondMessageErr(ctx.Respond, "Failed to delete tag: %s", err)
		}
		if ctx.User().ID != tag.OwnerID && ctx.Member().Permissions.Missing(discord.PermissionManageServer) {
			return common.RespondErrMessage(ctx.Respond, "You do not have permission to delete this tag.")
		}

		if err = b.DB.Delete(*ctx.GuildID(), name); err != nil {
			return common.RespondMessageErr(ctx.Respond, "Failed to delete tag: %s", err)
		}
		return common.Respond(ctx.Respond, "Tag deleted.")
	}
}

func infoTagHandler(b *butler.Butler) func(ctx *handler.CommandContext) error {
	return func(ctx *handler.CommandContext) error {
		data := ctx.SlashCommandInteractionData()
		name := formatTagName(data.String("name"))
		tag, err := b.DB.Get(*ctx.GuildID(), name)
		if err == sql.ErrNoRows {
			return common.Respondf(ctx.Respond, "Tag `%s` does not exist.", name)
		} else if err != nil {
			return common.RespondMessageErr(ctx.Respond, "Failed to get tag info: ", err)
		}
		return ctx.CreateMessage(discord.MessageCreate{
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

func listTagHandler(b *butler.Butler) func(ctx *handler.CommandContext) error {
	return func(ctx *handler.CommandContext) error {
		tags, err := b.DB.GetAll(*ctx.GuildID())
		if err != nil {
			return common.RespondMessageErr(ctx.Respond, "Failed to list tags: ", err)
		}
		if len(tags) == 0 {
			return common.Respond(ctx.Respond, "No tags found.")
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

		return b.Paginator.Create(ctx.Respond, &paginator.Paginator{
			PageFunc: func(page int, embed *discord.EmbedBuilder) {
				embed.SetDescription(pages[page])
			},
			MaxPages:        len(pages),
			ExpiryLastUsage: true,
			ID:              ctx.ID().String(),
		})
	}
}

func autoCompleteListTagHandler(b *butler.Butler, filterTags bool) func(ctx *handler.AutocompleteContext) error {
	return func(ctx *handler.AutocompleteContext) error {
		name := formatTagName(ctx.Data.String("name"))

		var (
			tags []db.Tag
			err  error
		)
		if filterTags && ctx.Member().Permissions.Missing(discord.PermissionManageServer) {
			tags, err = b.DB.GetAllUser(*ctx.GuildID(), ctx.User().ID)
		} else {
			tags, err = b.DB.GetAll(*ctx.GuildID())
		}

		if err != nil {
			return ctx.Result(nil)
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
		return ctx.Result(response)
	}
}

func formatTagName(name string) string {
	return strings.ToLower(name)
}
