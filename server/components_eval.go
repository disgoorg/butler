package server

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
)

func (s *Server) handleEvalRerun(event *handler.ComponentEvent) error {
	ownerID, hasOwner := interactionUserID(event.Message)
	if !hasOwner || ownerID != event.User().ID {
		return event.CreateMessage(discord.NewMessageCreate().WithContent("You can only rerun your own evals").WithEphemeral(true))
	}
	message, err := event.Client().Rest.GetMessage(event.Channel().ID(), snowflake.MustParse(event.Vars["message_id"]))
	if err != nil {
		return err
	}
	return s.eval(event.Client(), event.ComponentInteraction, event.Respond, message.Content, message.ID, true)
}

func (s *Server) handleEvalDelete(event *handler.ComponentEvent) error {
	ownerID, hasOwner := interactionUserID(event.Message)
	if !hasOwner || ownerID != event.User().ID {
		return event.CreateMessage(discord.NewMessageCreate().WithContent("You can only delete your own evals").WithEphemeral(true))
	}
	if err := event.DeferUpdateMessage(); err != nil {
		return err
	}
	return event.Client().Rest.DeleteMessage(event.Message.ChannelID, event.Message.ID)
}
