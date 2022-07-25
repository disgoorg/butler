package commands

import (
	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo-butler/common"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/json"
	"github.com/disgoorg/handler"
)

func TicketCommand(b *butler.Butler) handler.Command {
	return handler.Command{
		Create: discord.SlashCommandCreate{
			Name:         "close-ticket",
			Description:  "Closes the current ticket.",
			DMPermission: false,
		},
		CommandHandlers: map[string]handler.CommandHandler{
			"": modMailHandler(b),
		},
	}
}

func modMailHandler(b *butler.Butler) func(ctx *handler.CommandContext) error {
	return func(ctx *handler.CommandContext) error {
		b.ModMail.Mu.Lock()
		defer b.ModMail.Mu.Unlock()

		dmID, ok := b.ModMail.ThreadDMs[ctx.ChannelID()]
		if !ok {
			return common.RespondErrMessage(ctx.Respond, "No ticket found for this thread.")
		}
		delete(b.ModMail.ThreadDMs, ctx.ChannelID())
		delete(b.ModMail.DMThreads, dmID)

		if _, err := ctx.Client().Rest().CreateMessage(dmID, discord.MessageCreate{
			Embeds: []discord.Embed{
				{
					Author: &discord.EmbedAuthor{
						Name:    ctx.User().Tag(),
						IconURL: ctx.User().EffectiveAvatarURL(),
					},
					Description: "Ticket closed.",
					Color:       0xFF0000,
				},
			},
		}); err != nil {
			ctx.Client().Logger().Error("failed to close ticket in dm: ", err)
		}

		if err := ctx.CreateMessage(discord.MessageCreate{
			Embeds: []discord.Embed{
				{
					Description: "Ticket closed.",
					Color:       0x00FF00,
				},
			},
		}); err != nil {
			ctx.Client().Logger().Error("failed to respond to close ticket in channel: ", err)
		}

		_, err := ctx.Client().Rest().UpdateChannel(ctx.ChannelID(), discord.GuildThreadUpdate{
			Archived: json.NewPtr(true),
		})
		return err
	}
}
