package server

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/webhook"
	"github.com/disgoorg/snowflake/v2"

	"github.com/disgoorg/disgo-butler/server/database"
	"github.com/disgoorg/disgo-butler/server/database/sqlc"
)

func newModMail(cfg ModMailConfig, db *database.Database, threads map[snowflake.ID]snowflake.ID) *ModMail {
	modMail := &ModMail{
		db:               db,
		roleID:           cfg.RoleID,
		channelID:        cfg.ChannelID,
		webhookClient:    webhook.New(cfg.WebhookID, cfg.WebhookToken),
		DMThreads:        map[snowflake.ID]snowflake.ID{},
		ThreadDMs:        map[snowflake.ID]snowflake.ID{},
		dmMessageIDs:     map[snowflake.ID]snowflake.ID{},
		threadMessageIDs: map[snowflake.ID]snowflake.ID{},
	}
	for channelID, threadID := range threads {
		modMail.DMThreads[channelID] = threadID
		modMail.ThreadDMs[threadID] = channelID
	}

	modMail.ListenerAdapter = &events.ListenerAdapter{
		OnDMMessageCreate:        modMail.dmMessageCreateListener,
		OnDMMessageUpdate:        modMail.dmMessageUpdateListener,
		OnDMMessageDelete:        modMail.dmMessageDeleteListener,
		OnDMUserTypingStart:      modMail.dmUserTypingStartListener,
		OnGuildMessageCreate:     modMail.guildMessageCreateListener,
		OnGuildMessageUpdate:     modMail.guildMessageUpdateListener,
		OnGuildMessageDelete:     modMail.guildMessageDeleteListener,
		OnGuildMemberTypingStart: modMail.guildMemberTypingStartListener,
	}
	return modMail
}

var _ bot.EventListener = (*ModMail)(nil)

type ModMail struct {
	*events.ListenerAdapter
	db            *database.Database
	roleID        snowflake.ID
	channelID     snowflake.ID
	webhookClient *webhook.Client

	Mu sync.Mutex

	DMThreads map[snowflake.ID]snowflake.ID
	ThreadDMs map[snowflake.ID]snowflake.ID

	dmMessageIDs     map[snowflake.ID]snowflake.ID
	threadMessageIDs map[snowflake.ID]snowflake.ID
}

func generateEmbeds(message discord.Message) []discord.Embed {
	embeds := make([]discord.Embed, 0, len(message.Embeds)+1)
	embeds = append(embeds, discord.NewEmbed().
		WithAuthor(message.Author.Tag(), "", message.Author.EffectiveAvatarURL()).
		WithDescription(message.Content),
	)
	for i := range message.Embeds {
		if len(embeds) == 10 {
			break
		}
		embeds = append(embeds, message.Embeds[i])
	}
	return embeds
}

func filesFromAttachments(client *bot.Client, attachments []discord.Attachment) []*discord.File {
	var wg sync.WaitGroup
	files := make([]*discord.File, len(attachments))
	for i := range attachments {
		wg.Add(1)
		go func() {
			defer wg.Done()
			response, err := client.Rest.HTTPClient().Get(attachments[i].URL)
			if err != nil {
				client.Logger.Error("failed to get attachment", slog.Any("err", err))
				return
			}
			files[i] = discord.NewFile(attachments[i].Filename, "", response.Body)
		}()
	}
	wg.Wait()
	return files
}

func (m *ModMail) saveThread(channelID snowflake.ID, threadID snowflake.ID) {
	if err := m.db.CreateModMailThread(context.Background(), sqlc.CreateModMailThreadParams{ChannelID: channelID, ThreadID: threadID}); err != nil {
		slog.Error("failed to store mod mail thread", slog.Any("err", err))
	}
}

var roleMentions = &discord.AllowedMentions{
	Parse: []discord.AllowedMentionType{
		discord.AllowedMentionTypeRoles,
		discord.AllowedMentionTypeUsers,
		discord.AllowedMentionTypeEveryone,
	},
}

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
			newTicketMessage, err := event.Client().Rest.CreateMessage(event.ChannelID, discord.NewMessageCreate().
				WithEmbeds(discord.NewEmbed().WithDescription("Are you sure you want to open a ticket?")).
				AddActionRow(discord.NewSuccessButton("Yes", "yes"), discord.NewDangerButton("No", "no")),
			)
			if err != nil {
				event.Client().Logger.Error("failed to send new ticket message", slog.Any("err", err))
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			bot.WaitForEvent(event.Client(), ctx, func(component *events.ComponentInteractionCreate) bool {
				return component.Channel().ID() == event.ChannelID && component.Data.Type() == discord.ComponentTypeButton
			}, func(component *events.ComponentInteractionCreate) {
				if component.Data.CustomID() == "no" {
					accepted = false
					if err = component.UpdateMessage(discord.NewMessageUpdate().
						WithEmbeds(discord.NewEmbed().WithDescription("No Ticket created.").WithColor(0xFF0000)).
						ClearComponents(),
					); err != nil {
						event.Client().Logger.Error("failed to update new ticket message", slog.Any("err", err))
					}
					return
				}

				thread, err := event.Client().Rest.CreateThread(m.channelID, discord.GuildPublicThreadCreate{
					Name:                event.Message.Author.Tag(),
					AutoArchiveDuration: discord.AutoArchiveDuration1h,
				})
				if err != nil {
					accepted = false
					event.Client().Logger.Error("failed to create new thread", slog.Any("err", err))
					return
				}
				threadID = thread.ID()

				if _, err = m.webhookClient.CreateMessageInThread(discord.NewWebhookMessageCreate().
					WithContent(fmt.Sprintf("%s\nNew ticket opened by %s(`%s`)", discord.RoleMention(m.roleID), event.Message.Author.Tag(), event.Message.Author.ID)).
					WithAllowedMentions(roleMentions),
					threadID,
				); err != nil {
					event.Client().Logger.Error("failed to create new thread message", slog.Any("err", err))
				}

				m.Mu.Lock()
				defer m.Mu.Unlock()
				m.DMThreads[event.ChannelID] = threadID
				m.ThreadDMs[threadID] = event.ChannelID
				m.saveThread(event.ChannelID, threadID)
				if err = component.UpdateMessage(discord.NewMessageUpdate().
					WithEmbeds(discord.NewEmbed().WithDescription("New Ticket created.").WithColor(0x00FF00)).
					ClearComponents(),
				); err != nil {
					event.Client().Logger.Error("failed to update new ticket message", slog.Any("err", err))
				}
			}, func() {
				accepted = false
				if _, err = event.Client().Rest.UpdateMessage(event.ChannelID, newTicketMessage.ID, discord.NewMessageUpdate().
					WithEmbeds(discord.NewEmbed().WithDescription("Ticket creation timed out.").WithColor(0xFF0000)).
					ClearComponents(),
				); err != nil {
					event.Client().Logger.Error("failed to update new ticket message", slog.Any("err", err))
				}
			})
			if !accepted {
				return
			}
		}

		message, err := m.webhookClient.CreateMessageInThread(discord.NewWebhookMessageCreate().
			WithContent(event.Message.Content).
			WithUsername(event.Message.Author.Username).
			WithAvatarURL(event.Message.Author.EffectiveAvatarURL()).
			WithEmbeds(event.Message.Embeds...).
			WithFiles(filesFromAttachments(event.Client(), event.Message.Attachments)...),
			threadID,
		)
		if err != nil {
			event.Client().Logger.Error("failed to create thread message", slog.Any("err", err))
			return
		}
		m.Mu.Lock()
		defer m.Mu.Unlock()
		m.threadMessageIDs[event.Message.ID] = message.ID
	}()
}
