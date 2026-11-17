package server

import (
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
)

var ticketCommand = discord.SlashCommandCreate{
	Name:        "close-ticket",
	Description: "Closes the current ticket.",
	Contexts: []discord.InteractionContextType{
		discord.InteractionContextTypeGuild,
		discord.InteractionContextTypeBotDM,
	},
}

func (s *Server) handleCloseTicket(event *handler.CommandEvent) error {
	s.ModMail.Mu.Lock()
	defer s.ModMail.Mu.Unlock()

	channelID := event.Channel().ID()
	var dmID, threadID snowflake.ID
	if event.GuildID() == nil {
		var ok bool
		threadID, ok = s.ModMail.DMThreads[channelID]
		if !ok {
			return respondErrMessage(event.Respond, "No ticket found for this thread.")
		}
		dmID = channelID
	} else {
		var ok bool
		dmID, ok = s.ModMail.ThreadDMs[channelID]
		if !ok {
			return respondErrMessage(event.Respond, "No ticket found for this thread.")
		}
		threadID = channelID
	}
	delete(s.ModMail.ThreadDMs, threadID)
	delete(s.ModMail.DMThreads, dmID)
	if err := s.DB.DeleteModMailThread(event.Ctx, channelID); err != nil {
		event.Client().Logger.Error("failed to delete mod mail thread", slog.Any("err", err))
	}

	archived := true
	if _, err := event.Client().Rest.CreateMessage(dmID, discord.NewMessageCreate().WithEmbeds(discord.NewEmbed().
		WithAuthor(event.User().Tag(), "", event.User().EffectiveAvatarURL()).
		WithDescription("Ticket closed.").
		WithColor(0xFF0000),
	)); err != nil {
		event.Client().Logger.Error("failed to close ticket in dm", slog.Any("err", err))
	}

	if err := event.CreateMessage(discord.NewMessageCreate().WithEmbeds(discord.NewEmbed().
		WithDescription("Ticket closed.").
		WithColor(0x00FF00),
	)); err != nil {
		event.Client().Logger.Error("failed to respond to close ticket", slog.Any("err", err))
	}

	_, err := event.Client().Rest.UpdateChannel(threadID, discord.GuildThreadUpdate{Archived: &archived})
	return err
}
