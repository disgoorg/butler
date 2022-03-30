package commands

import (
	"context"

	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo-butler/models"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func createTagHandler(b *butler.Butler, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()

	var message string
	if _, err := b.DB.NewInsert().Model(&models.Tag{
		GuildID: *e.GuildID(),
		Name:    data.String("name"),
		Content: data.String("content"),
	}).Exec(context.TODO()); err != nil {
		message = "Failed to create tag: " + err.Error()
	} else {
		message = "Tag created!"
	}
	return e.CreateMessage(discord.MessageCreate{
		Content: message,
	})
}

func deleteTagHandler(b *butler.Butler, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	var message string
	if _, err := b.DB.NewDelete().Model(&models.Tag{
		GuildID: *e.GuildID(),
		Name:    data.String("name"),
	}).Exec(context.TODO()); err != nil {
		message = "Failed to delete tag: " + err.Error()
	} else {
		message = "Tag deleted!"
	}
	return e.CreateMessage(discord.MessageCreate{
		Content: message,
	})
}

func editTagHandler(b *butler.Butler, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	var message string
	if _, err := b.DB.NewUpdate().Model(&models.Tag{
		GuildID: *e.GuildID(),
		Name:    data.String("name"),
		Content: data.String("content"),
	}).Exec(context.TODO()); err != nil {
		message = "Failed to edit tag: " + err.Error()
	} else {
		message = "Tag edited!"
	}
	return e.CreateMessage(discord.MessageCreate{
		Content: message,
	})
}
