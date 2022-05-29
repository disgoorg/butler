package mod_mail

import (
	"sync"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/webhook"
	"github.com/disgoorg/snowflake/v2"
)

func New(config Config) *ModMail {
	modMail := &ModMail{

		channelID:        config.ChannelID,
		webhookClient:    webhook.New(config.WebhookID, config.WebhookToken),
		dmThreads:        map[snowflake.ID]snowflake.ID{},
		threadDMs:        map[snowflake.ID]snowflake.ID{},
		dmMessageIDs:     map[snowflake.ID]snowflake.ID{},
		threadMessageIDs: map[snowflake.ID]snowflake.ID{},
	}
	for _, thread := range config.Threads {
		modMail.dmThreads[thread.ChannelID] = thread.ThreadID
		modMail.threadDMs[thread.ThreadID] = thread.ChannelID
	}

	modMail.ListenerAdapter = events.ListenerAdapter{
		OnDMMessageCreate:   modMail.dmMessageCreateListener,
		OnDMMessageUpdate:   modMail.dmMessageUpdateListener,
		OnDMMessageDelete:   modMail.dmMessageDeleteListener,
		OnDMUserTypingStart: modMail.dmUserTypingStartListener,

		OnGuildMessageCreate:     modMail.guildMessageCreateListener,
		OnGuildMessageUpdate:     modMail.guildMessageUpdateListener,
		OnGuildMessageDelete:     modMail.guildMessageDeleteListener,
		OnGuildMemberTypingStart: modMail.guildMemberTypingStartListener,
	}

	return modMail
}

var _ bot.EventListener = (*ModMail)(nil)

type ModMail struct {
	events.ListenerAdapter
	channelID     snowflake.ID
	webhookClient webhook.Client

	mu sync.Mutex

	// DMChannelID -> ThreadID
	dmThreads map[snowflake.ID]snowflake.ID
	// ThreadID -> DMChannelID
	threadDMs map[snowflake.ID]snowflake.ID

	// DMMessageID -> ThreadMessageID
	dmMessageIDs map[snowflake.ID]snowflake.ID
	// ThreadMessageID -> DMMessageID
	threadMessageIDs map[snowflake.ID]snowflake.ID
}

func (m *ModMail) Close() []Thread {
	m.mu.Lock()
	defer m.mu.Unlock()

	threads := make([]Thread, len(m.dmThreads))
	var i int
	for dmID, threadID := range m.dmThreads {
		threads[i] = Thread{
			ChannelID: dmID,
			ThreadID:  threadID,
		}
		i++
	}
	return threads
}

func generateEmbeds(message discord.Message) []discord.Embed {
	embeds := make([]discord.Embed, len(message.Embeds)+1)
	embeds[0] = discord.Embed{
		Author: &discord.EmbedAuthor{
			Name:    message.Author.Tag(),
			IconURL: message.Author.EffectiveAvatarURL(),
		},
		Description: message.Content,
	}

	for i := range message.Embeds {
		if len(embeds) == 10 {
			break
		}
		embeds[i+1] = message.Embeds[i]
	}
	return embeds
}

func filesFromAttachments(client bot.Client, attachments []discord.Attachment) []*discord.File {
	var wg sync.WaitGroup
	files := make([]*discord.File, len(attachments))
	for ii := range attachments {
		wg.Add(1)
		i := ii
		go func() {
			defer wg.Done()
			rs, err := client.Rest().HTTPClient().Get(attachments[i].URL)
			if err != nil {
				client.Logger().Error("failed to get attachment: ", err)
				return
			}
			files[i] = discord.NewFile(attachments[i].Filename, "", rs.Body)
		}()
	}
	wg.Wait()
	return files
}

type Config struct {
	ChannelID    snowflake.ID `json:"channel_id"`
	WebhookID    snowflake.ID `json:"webhook_id"`
	WebhookToken string       `json:"webhook_token"`
	Threads      []Thread     `json:"threads"`
}

type Thread struct {
	ThreadID  snowflake.ID `json:"thread_id"`
	ChannelID snowflake.ID `json:"channel_id"`
}
