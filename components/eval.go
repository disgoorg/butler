package components

import (
	"fmt"

	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo-butler/commands"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
)

func HandleEvalRerunAction(b *butler.Butler) handler.ComponentHandler {
	return func(e *handler.ComponentEvent) error {
		fmt.Printf("rerun eval: %#v\n", e.Message.Interaction)
		if e.Message.Interaction.User.ID != e.User().ID {
			return e.CreateMessage(discord.MessageCreate{Content: "You can only rerun your own evals", Flags: discord.MessageFlagEphemeral})
		}
		message, err := e.Client().Rest().GetMessage(e.ChannelID(), snowflake.MustParse(e.Variables["message_id"]))
		if err != nil {
			return err
		}

		fmt.Printf("HandleEvalRerunAction: %#v\n", e.Message)
		return commands.Eval(b, e.Client(), e.BaseInteraction, e.Respond, message.Content, message.ID, true)
	}
}

func HandleEvalDeleteAction(e *handler.ComponentEvent) error {
	if e.Message.Interaction.User.ID != e.User().ID {
		return e.CreateMessage(discord.MessageCreate{Content: "You can only delete your own evals", Flags: discord.MessageFlagEphemeral})
	}
	if err := e.DeferUpdateMessage(); err != nil {
		return err
	}
	return e.Client().Rest().DeleteMessage(e.Message.ChannelID, e.Message.ID)
}
