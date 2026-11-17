package server

import (
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func (m *ModMail) guildMessageCreateListener(event *events.GuildMessageCreate) {
	if event.Message.WebhookID != nil {
		return
	}

	m.Mu.Lock()
	defer m.Mu.Unlock()
	dmID, ok := m.ThreadDMs[event.ChannelID]
	if !ok {
		return
	}

	message, err := event.Client().Rest.CreateMessage(dmID, discord.NewMessageCreate().
		WithEmbeds(generateEmbeds(event.Message)...).
		WithFiles(filesFromAttachments(event.Client(), event.Message.Attachments)...),
	)
	if err != nil {
		event.Client().Logger.Error("failed to create dm message", slog.Any("err", err))
		return
	}
	m.dmMessageIDs[event.Message.ID] = message.ID
}

func (m *ModMail) guildMessageUpdateListener(event *events.GuildMessageUpdate) {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	dmMessageID, ok := m.dmMessageIDs[event.Message.ID]
	if !ok {
		return
	}
	dmChannelID := m.ThreadDMs[event.ChannelID]
	if _, err := event.Client().Rest.UpdateMessage(dmChannelID, dmMessageID, discord.NewMessageUpdate().
		WithEmbeds(generateEmbeds(event.Message)...).
		WithFiles(filesFromAttachments(event.Client(), event.Message.Attachments)...),
	); err != nil {
		event.Client().Logger.Error("failed to update dm message", slog.Any("err", err))
	}
}

func (m *ModMail) guildMessageDeleteListener(event *events.GuildMessageDelete) {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	dmMessageID, ok := m.dmMessageIDs[event.MessageID]
	if !ok {
		return
	}
	delete(m.dmMessageIDs, event.MessageID)
	dmChannelID := m.ThreadDMs[event.ChannelID]
	if err := event.Client().Rest.DeleteMessage(dmChannelID, dmMessageID); err != nil {
		event.Client().Logger.Error("failed to delete dm message", slog.Any("err", err))
	}
}

func (m *ModMail) guildMemberTypingStartListener(event *events.GuildMemberTypingStart) {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	dmChannelID, ok := m.ThreadDMs[event.ChannelID]
	if !ok {
		return
	}
	if err := event.Client().Rest.SendTyping(dmChannelID); err != nil {
		event.Client().Logger.Error("failed to send dm typing", slog.Any("err", err))
	}
}
