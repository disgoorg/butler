package mod_mail

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func (m *ModMail) dmMessageCreateListener(event *events.DMMessageCreate) {
	if event.Message.Author.ID == event.Client().ID() {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	threadID, ok := m.dmThreads[event.ChannelID]
	if !ok {
		thread, err := event.Client().Rest().CreateThread(m.channelID, discord.GuildPublicThreadCreate{
			Name:                event.Message.Author.Tag(),
			AutoArchiveDuration: discord.AutoArchiveDuration1h,
		})
		if err != nil {
			event.Client().Logger().Error("failed to create new thread: ", err)
			return
		}
		threadID = thread.ID()
		m.dmThreads[event.ChannelID] = thread.ID()
		m.threadDMs[thread.ID()] = event.ChannelID
	}
	webhookMessageCreate := discord.WebhookMessageCreate{
		Content:   event.Message.Content,
		Username:  event.Message.Author.Username,
		AvatarURL: event.Message.Author.EffectiveAvatarURL(),
		Embeds:    event.Message.Embeds,
		Files:     filesFromAttachments(event.Client(), event.Message.Attachments),
	}

	message, err := m.webhookClient.CreateMessageInThread(webhookMessageCreate, threadID)
	if err != nil {
		event.Client().Logger().Error("failed to create thread message: ", err)
		return
	}
	m.threadMessageIDs[event.Message.ID] = message.ID

}

func (m *ModMail) dmMessageUpdateListener(event *events.DMMessageUpdate) {
	m.mu.Lock()
	defer m.mu.Unlock()

	webhookMessageID, ok := m.threadMessageIDs[event.Message.ID]
	if !ok {
		return
	}
	webhookMessageUpdate := discord.WebhookMessageUpdate{
		Content: &event.Message.Content,
		Embeds:  &event.Message.Embeds,
		Files:   filesFromAttachments(event.Client(), event.Message.Attachments),
	}
	threadID := m.dmThreads[event.ChannelID]
	_, err := m.webhookClient.UpdateMessageInThread(webhookMessageID, webhookMessageUpdate, threadID)
	if err != nil {
		event.Client().Logger().Error("failed to update thread message: ", err)
		return
	}

}

func (m *ModMail) dmMessageDeleteListener(event *events.DMMessageDelete) {
	m.mu.Lock()
	defer m.mu.Unlock()

	webhookMessageID, ok := m.threadMessageIDs[event.MessageID]
	if !ok {
		return
	}
	delete(m.threadMessageIDs, event.Message.ID)
	if err := m.webhookClient.DeleteMessage(webhookMessageID); err != nil {
		event.Client().Logger().Error("failed to delete thread message: ", err)
		return
	}

}

func (m *ModMail) dmUserTypingStartListener(event *events.DMUserTypingStart) {
	m.mu.Lock()
	defer m.mu.Unlock()

	threadID, ok := m.dmThreads[event.ChannelID]
	if !ok {
		return
	}
	if err := event.Client().Rest().SendTyping(threadID); err != nil {
		event.Client().Logger().Error("failed to send thread typing: ", err)
		return
	}

}
