package server

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/paginator"
	"github.com/lithammer/fuzzysearch/fuzzy"

	"github.com/disgoorg/disgo-butler/server/database"
	"github.com/disgoorg/disgo-butler/server/database/sqlc"
)

var tagCommand = discord.SlashCommandCreate{
	Name:        "tag",
	Description: "Let's you display a tag",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionString{
			Name:         "name",
			Description:  "The name of the tag to display",
			Required:     true,
			Autocomplete: true,
		},
	},
}

func (s *Server) handleTag(event *handler.CommandEvent) error {
	content, err := s.DB.GetAndIncrementTag(event.Ctx, sqlc.GetAndIncrementTagParams{
		GuildID: *event.GuildID(),
		Name:    event.SlashCommandInteractionData().String("name"),
	})
	if database.IsNotFound(err) {
		return respondErrMessage(event.Respond, "Tag not found")
	} else if err != nil {
		return respondErr(event.Respond, err)
	}
	return event.CreateMessage(discord.NewMessageCreate().WithContent(content))
}

var tagsCommand = discord.SlashCommandCreate{
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
}

func (s *Server) handleCreateTag(event *handler.CommandEvent) error {
	data := event.SlashCommandInteractionData()
	name := formatTagName(data.String("name"))

	if _, err := s.DB.GetTag(event.Ctx, sqlc.GetTagParams{GuildID: *event.GuildID(), Name: name}); err == nil {
		return respondErrMessage(event.Respond, "Tag already exists.")
	} else if err != nil && !database.IsNotFound(err) {
		return respondMessageErr(event.Respond, "Failed to edit tag: %s", err)
	}

	if err := s.DB.CreateTag(event.Ctx, sqlc.CreateTagParams{
		GuildID: *event.GuildID(),
		OwnerID: event.User().ID,
		Name:    name,
		Content: data.String("content"),
	}); err != nil {
		return respondMessageErr(event.Respond, "Failed to create tag: %s", err)
	}
	return respond(event.Respond, "Tag created!")
}

func (s *Server) handleEditTag(event *handler.CommandEvent) error {
	data := event.SlashCommandInteractionData()
	name := formatTagName(data.String("name"))

	tag, err := s.DB.GetTag(event.Ctx, sqlc.GetTagParams{GuildID: *event.GuildID(), Name: name})
	if database.IsNotFound(err) {
		return respondErrMessage(event.Respond, "Tag not found.")
	} else if err != nil {
		return respondMessageErr(event.Respond, "Failed to edit tag: %s", err)
	}
	if event.User().ID != tag.OwnerID && missingPermission(event.Member(), discord.PermissionManageGuild) {
		return respondErrMessage(event.Respond, "You do not have permission to edit this tag.")
	}

	if err = s.DB.UpdateTag(event.Ctx, sqlc.UpdateTagParams{GuildID: *event.GuildID(), Name: name, Content: data.String("content")}); err != nil {
		return respondMessageErr(event.Respond, "Failed to edit tag: %s", err)
	}
	return respond(event.Respond, "Tag edited.")
}

func (s *Server) handleDeleteTag(event *handler.CommandEvent) error {
	data := event.SlashCommandInteractionData()
	name := formatTagName(data.String("name"))

	tag, err := s.DB.GetTag(event.Ctx, sqlc.GetTagParams{GuildID: *event.GuildID(), Name: name})
	if database.IsNotFound(err) {
		return respondErrMessage(event.Respond, "Tag not found.")
	} else if err != nil {
		return respondMessageErr(event.Respond, "Failed to delete tag: %s", err)
	}
	if event.User().ID != tag.OwnerID && missingPermission(event.Member(), discord.PermissionManageGuild) {
		return respondErrMessage(event.Respond, "You do not have permission to delete this tag.")
	}

	if err = s.DB.DeleteTag(event.Ctx, sqlc.DeleteTagParams{GuildID: *event.GuildID(), Name: name}); err != nil {
		return respondMessageErr(event.Respond, "Failed to delete tag: %s", err)
	}
	return respond(event.Respond, "Tag deleted.")
}

func (s *Server) handleTagInfo(event *handler.CommandEvent) error {
	data := event.SlashCommandInteractionData()
	name := formatTagName(data.String("name"))
	tag, err := s.DB.GetTag(event.Ctx, sqlc.GetTagParams{GuildID: *event.GuildID(), Name: name})
	if database.IsNotFound(err) {
		return respondf(event.Respond, "Tag `%s` does not exist.", name)
	} else if err != nil {
		return respondMessageErr(event.Respond, "Failed to get tag info: %s", err)
	}
	return event.CreateMessage(discord.NewMessageCreate().WithEmbeds(discord.Embed{
		Title:       fmt.Sprintf("Tag `%s`", name),
		Description: tag.Content,
		Fields: []discord.EmbedField{
			{Name: "Created by", Value: discord.UserMention(tag.OwnerID)},
			{Name: "Uses", Value: strconv.Itoa(int(tag.Uses))},
			{Name: "Created at", Value: fmt.Sprintf("%s (%s)", discord.NewTimestamp(discord.TimestampStyleNone, tag.CreatedAt), discord.NewTimestamp(discord.TimestampStyleRelative, tag.CreatedAt))},
		},
	}))
}

func (s *Server) handleListTags(event *handler.CommandEvent) error {
	tags, err := s.DB.ListTags(event.Ctx, *event.GuildID())
	if err != nil {
		return respondMessageErr(event.Respond, "Failed to list tags: %s", err)
	}
	if len(tags) == 0 {
		return respond(event.Respond, "No tags found.")
	}

	var pages []string
	curPage := ""
	for _, tag := range tags {
		line := fmt.Sprintf("**%s** - %s\n", tag.Name, discord.UserMention(tag.OwnerID))
		if len(curPage)+len(line) > 2000 {
			pages = append(pages, curPage)
			curPage = ""
		}
		curPage += line
	}
	if len(curPage) > 0 {
		pages = append(pages, curPage)
	}

	return s.Paginator.Create(event.Respond, paginator.Pages{
		ID: event.ID().String(),
		PageFunc: func(page int, embed discord.Embed) discord.Embed {
			embed.Description = pages[page]
			return embed
		},
		Pages:      len(pages),
		ExpireMode: paginator.ExpireModeAfterLastUsage,
	}, false)
}

func (s *Server) handleTagListAutocomplete(filterTags bool) handler.AutocompleteHandler {
	return func(event *handler.AutocompleteEvent) error {
		name := formatTagName(event.Data.String("name"))

		names, err := s.tagNames(event, filterTags)
		if err != nil {
			return event.AutocompleteResult(nil)
		}

		var response []discord.AutocompleteChoice
		for _, option := range fuzzy.FindFold(name, names) {
			if len(response) >= 25 {
				break
			}
			response = append(response, discord.AutocompleteChoiceString{Name: option, Value: option})
		}
		return event.AutocompleteResult(response)
	}
}

func (s *Server) tagNames(event *handler.AutocompleteEvent, filterTags bool) ([]string, error) {
	if filterTags && missingPermission(event.Member(), discord.PermissionManageGuild) {
		tags, err := s.DB.ListTagsByOwner(event.Ctx, sqlc.ListTagsByOwnerParams{GuildID: *event.GuildID(), OwnerID: event.User().ID})
		if err != nil {
			return nil, err
		}
		names := make([]string, len(tags))
		for i, tag := range tags {
			names[i] = tag.Name
		}
		return names, nil
	}

	tags, err := s.DB.ListTags(event.Ctx, *event.GuildID())
	if err != nil {
		return nil, err
	}
	names := make([]string, len(tags))
	for i, tag := range tags {
		names[i] = tag.Name
	}
	return names, nil
}

func formatTagName(name string) string {
	return strings.ToLower(name)
}

func missingPermission(member *discord.ResolvedMember, permission discord.Permissions) bool {
	if member == nil {
		return true
	}
	return member.Permissions.Missing(permission)
}
