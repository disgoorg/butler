package server

import (
	"context"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
)

func (s *Server) handleDocsAction(event *handler.ComponentEvent) error {
	action := event.StringSelectMenuInteractionData().Values[0]
	ownerID, hasOwner := interactionUserID(event.Message)
	if action == "delete" {
		if (!hasOwner || ownerID != event.User().ID) && missingPermission(event.Member(), discord.PermissionManageMessages) {
			return respondErrMessage(event.Respond, "You don't have permission to delete this message.")
		}
		_ = event.DeferUpdateMessage()
		return event.Client().Rest.DeleteInteractionResponse(event.ApplicationID(), event.Token())
	}
	values := strings.SplitN(event.Message.Embeds[0].Title, ": ", 2)
	pkg, err := s.DocClient.Search(context.Background(), values[0])
	if err != nil {
		return respondErrMessagef(event.Respond, "Error while fetching package: %s", err)
	}

	var (
		expandSignature bool
		expandComment   bool
		expandMethods   bool
		expandExamples  bool
	)
	if strings.HasPrefix(action, "expand:") {
		switch strings.TrimPrefix(action, "expand:") {
		case "signature":
			expandSignature = true
		case "methods":
			expandMethods = true
		case "comment":
			expandComment = true
		case "examples":
			expandExamples = true
		}
	} else if !strings.HasPrefix(action, "collapse:") {
		return respondErrMessagef(event.Respond, "Unknown action: %s", action)
	}

	var query string
	if len(values) > 1 {
		query = values[1]
	}
	embed, selectMenu := getDocsEmbed(pkg, query, expandSignature, expandComment, expandMethods, expandExamples)
	if (!hasOwner || ownerID != event.User().ID) && missingPermission(event.Member(), discord.PermissionManageMessages) {
		return event.CreateMessage(discord.NewMessageCreate().WithEmbeds(embed).WithEphemeral(true))
	}
	return event.UpdateMessage(discord.NewMessageUpdate().WithEmbeds(embed).AddActionRow(selectMenu))
}

func interactionUserID(message discord.Message) (snowflake.ID, bool) {
	if message.InteractionMetadata != nil {
		return message.InteractionMetadata.User.ID, true
	}
	if message.Interaction != nil {
		return message.Interaction.User.ID, true
	}
	return 0, false
}
