package commands

import (
	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo-butler/common"
	"github.com/disgoorg/disgo-butler/mod_mail"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/json"
)

var TicketCommand = func(m *mod_mail.ModMail) butler.Command {
	return butler.Command{
		Create: discord.SlashCommandCreate{
			CommandName:  "close-ticket",
			Description:  "Closes the current ticket.",
			DMPermission: false,
		},
		CommandHandlers: map[string]butler.HandleFunc{
			"": func(b *butler.Butler, e *events.ApplicationCommandInteractionCreate) error {
				m.Mu.Lock()
				defer m.Mu.Unlock()

				dmID, ok := m.ThreadDMs[e.ChannelID()]
				if !ok {
					return common.RespondErrMessage(e.Respond, "No ticket found for this thread.")
				}
				delete(m.ThreadDMs, e.ChannelID())
				delete(m.DMThreads, dmID)

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

			},
		},
	}
}
