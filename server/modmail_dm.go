package server

import (
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func (m *ModMail) dmMessageUpdateListener(event *events.DMMessageUpdate) {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	webhookMessageID, ok := m.threadMessageIDs[event.Message.ID]
	if !ok {
		return
	}
	threadID := m.DMThreads[event.ChannelID]
	if _, err := m.webhookClient.UpdateMessageInThread(webhookMessageID, discord.NewWebhookMessageUpdate().
		WithContent(event.Message.Content).
		WithEmbeds(event.Message.Embeds...).
		WithFiles(filesFromAttachments(event.Client(), event.Message.Attachments)...),
		threadID,
	); err != nil {
		event.Client().Logger.Error("failed to update thread message", slog.Any("err", err))
	}
}

func (m *ModMail) dmMessageDeleteListener(event *events.DMMessageDelete) {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	webhookMessageID, ok := m.threadMessageIDs[event.MessageID]
	if !ok {
		return
	}
	delete(m.threadMessageIDs, event.MessageID)
	if err := m.webhookClient.DeleteMessage(webhookMessageID); err != nil {
		event.Client().Logger.Error("failed to delete thread message", slog.Any("err", err))
	}
}

func (m *ModMail) dmUserTypingStartListener(event *events.DMUserTypingStart) {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	threadID, ok := m.DMThreads[event.ChannelID]
	if !ok {
		return
	}
	if err := event.Client().Rest.SendTyping(threadID); err != nil {
		event.Client().Logger.Error("failed to send thread typing", slog.Any("err", err))
	}
}
