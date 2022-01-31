package commands

import (
	"context"

	"github.com/DisgoOrg/disgo-butler/butler"
	"github.com/DisgoOrg/disgo-butler/models"
	"github.com/DisgoOrg/disgo/core/events"
	"github.com/DisgoOrg/disgo/discord"
)

func createTagHandler(b *butler.Butler, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()

	var message string
	if _, err := b.DB.NewInsert().Model(&models.Tag{
		GuildID: *e.GuildID,
		Name:    *data.Options.String("name"),
		Content: *data.Options.String("content"),
	}).Exec(context.TODO()); err != nil {
		message = "Failed to create tag: " + err.Error()
	} else {
		message = "Tag created!"
	}
	return e.Create(discord.MessageCreate{
		Content: message,
	})
}

func deleteTagHandler(b *butler.Butler, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	var message string
	if _, err := b.DB.NewDelete().Model(&models.Tag{
		GuildID: *e.GuildID,
		Name:    *data.Options.String("name"),
	}).Exec(context.TODO()); err != nil {
		message = "Failed to delete tag: " + err.Error()
	} else {
		message = "Tag deleted!"
	}
	return e.Create(discord.MessageCreate{
		Content: message,
	})
}

func editTagHandler(b *butler.Butler, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	var message string
	if _, err := b.DB.NewUpdate().Model(&models.Tag{
		GuildID: *e.GuildID,
		Name:    *data.Options.String("name"),
		Content: *data.Options.String("content"),
	}).Exec(context.TODO()); err != nil {
		message = "Failed to edit tag: " + err.Error()
	} else {
		message = "Tag edited!"
	}
	return e.Create(discord.MessageCreate{
		Content: message,
	})
}
