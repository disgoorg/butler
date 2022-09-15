package mod_mail

import (
	"context"
	"fmt"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func (m *ModMail) dmMessageCreateListener(event *events.DMMessageCreate) {
	if event.Message.Author.ID == event.Client().ID() {
		return
	}

	go func() {
		accepted := true
		m.Mu.Lock()
		threadID, ok := m.DMThreads[event.ChannelID]
		m.Mu.Unlock()
		if !ok {
			newTicketMessage, err := event.Client().Rest().CreateMessage(event.ChannelID, discord.NewMessageCreateBuilder().
				SetEmbeds(discord.NewEmbedBuilder().
					SetDescription("Are you sure you want to open a ticket?").
					Build(),
				).
				AddActionRow(discord.NewSuccessButton("Yes", "yes"), discord.NewDangerButton("No", "no")).
				Build(),
			)
			if err != nil {
				event.Client().Logger().Error("failed to send new ticket message: ", err)
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			bot.WaitForEvent(event.Client(), ctx, func(e *events.ComponentInteractionCreate) bool {
				return e.ChannelID() == event.ChannelID && e.Data.Type() == discord.ComponentTypeButton
			}, func(e *events.ComponentInteractionCreate) {
				if e.Data.CustomID() == "no" {
					accepted = false
					if err = e.UpdateMessage(discord.MessageUpdate{
						Embeds: &[]discord.Embed{
							{
								Description: "No Ticket created.",
								Color:       0xFF0000,
							},
						},
						Components: &[]discord.ContainerComponent{},
					}); err != nil {
						event.Client().Logger().Error("failed to update new ticket message: ", err)
					}
					return
				}

				thread, err := event.Client().Rest().CreateThread(m.channelID, discord.GuildPublicThreadCreate{
					Name:                event.Message.Author.Tag(),
					AutoArchiveDuration: discord.AutoArchiveDuration1h,
				})
				if err != nil {
					event.Client().Logger().Error("failed to create new thread: ", err)
					return
				}

				threadID = thread.ID()

				if _, err = m.webhookClient.CreateMessageInThread(discord.WebhookMessageCreate{
					Content:         fmt.Sprintf("%s\nNew ticket opened by %s(`%s`)", discord.RoleMention(m.roleID), event.Message.Author.Tag(), event.Message.Author.ID),
					AllowedMentions: &discord.DefaultAllowedMentions,
				}, threadID); err != nil {
					event.Client().Logger().Error("failed to create new thread message: ", err)
				}

				m.Mu.Lock()
				defer m.Mu.Unlock()
				m.DMThreads[event.ChannelID] = threadID
				m.ThreadDMs[threadID] = event.ChannelID
				if err = e.UpdateMessage(discord.MessageUpdate{
					Embeds: &[]discord.Embed{
						{
							Description: "New Ticket created.",
							Color:       0x00FF00,
						},
					},
					Components: &[]discord.ContainerComponent{},
				}); err != nil {
					event.Client().Logger().Error("failed to update new ticket message: ", err)
				}
			}, func() {
				accepted = false
				if _, err = event.Client().Rest().UpdateMessage(event.ChannelID, newTicketMessage.ID, discord.MessageUpdate{
					Embeds: &[]discord.Embed{
						{
							Description: "Ticket creation timed out.",
							Color:       0xFF0000,
						},
					},
					Components: &[]discord.ContainerComponent{},
				}); err != nil {
					event.Client().Logger().Error("failed to update new ticket message: ", err)
				}
			})
			if !accepted {
				return
			}
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
		m.Mu.Lock()
		defer m.Mu.Unlock()
		m.threadMessageIDs[event.Message.ID] = message.ID
	}()
}

func (m *ModMail) dmMessageUpdateListener(event *events.DMMessageUpdate) {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	webhookMessageID, ok := m.threadMessageIDs[event.Message.ID]
	if !ok {
		return
	}
	webhookMessageUpdate := discord.WebhookMessageUpdate{
		Content: &event.Message.Content,
		Embeds:  &event.Message.Embeds,
		Files:   filesFromAttachments(event.Client(), event.Message.Attachments),
	}
	threadID := m.DMThreads[event.ChannelID]
	_, err := m.webhookClient.UpdateMessageInThread(webhookMessageID, webhookMessageUpdate, threadID)
	if err != nil {
		event.Client().Logger().Error("failed to update thread message: ", err)
		return
	}

}

func (m *ModMail) dmMessageDeleteListener(event *events.DMMessageDelete) {
	m.Mu.Lock()
	defer m.Mu.Unlock()

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
	m.Mu.Lock()
	defer m.Mu.Unlock()

	threadID, ok := m.DMThreads[event.ChannelID]
	if !ok {
		return
	}
	if err := event.Client().Rest().SendTyping(threadID); err != nil {
		event.Client().Logger().Error("failed to send thread typing: ", err)
		return
	}

}
