package commands

import (
	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo-butler/common"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/json"
	"github.com/disgoorg/handler"
)

func TicketCommand(b *butler.Butler) handler.Command {
	return handler.Command{
		Create: discord.SlashCommandCreate{
			Name:         "close-ticket",
			Description:  "Closes the current ticket.",
			DMPermission: json.NewPtr(true),
		},
		CommandHandlers: map[string]handler.CommandHandler{
			"": modMailHandler(b),
		},
	}
}

func modMailHandler(b *butler.Butler) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		b.ModMail.Mu.Lock()
		defer b.ModMail.Mu.Unlock()

		if e.GuildID() == nil {
			threadID, ok := b.ModMail.DMThreads[e.ChannelID()]
			if !ok {
				return common.RespondErrMessage(e.Respond, "No ticket found for this thread.")
			}
			delete(b.ModMail.ThreadDMs, e.ChannelID())
			delete(b.ModMail.DMThreads, dmID)
		}

		dmID, ok := b.ModMail.ThreadDMs[e.ChannelID()]
		if !ok {
			return common.RespondErrMessage(e.Respond, "No ticket found for this thread.")
		}
		delete(b.ModMail.ThreadDMs, e.ChannelID())
		delete(b.ModMail.DMThreads, dmID)

		if _, err := e.Client().Rest().CreateMessage(dmID, discord.MessageCreate{
			Embeds: []discord.Embed{
				{
					Author: &discord.EmbedAuthor{
						Name:    e.User().Tag(),
						IconURL: e.User().EffectiveAvatarURL(),
					},
					Description: "Ticket closed.",
					Color:       0xFF0000,
				},
			},
		}); err != nil {
			e.Client().Logger().Error("failed to close ticket in dm: ", err)
		}

		if err := e.CreateMessage(discord.MessageCreate{
			Embeds: []discord.Embed{
				{
					Description: "Ticket closed.",
					Color:       0x00FF00,
				},
			},
		}); err != nil {
			e.Client().Logger().Error("failed to respond to close ticket in channel: ", err)
		}

		_, err := e.Client().Rest().UpdateChannel(e.ChannelID(), discord.GuildThreadUpdate{
			Archived: json.NewPtr(true),
		})
		return err
	}
}
