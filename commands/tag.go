package commands

import (
	"database/sql"

	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo-butler/common"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/handler"
)

func TagCommand(b *butler.Butler) handler.Command {
	return handler.Command{
		Create: discord.SlashCommandCreate{
			CommandName: "tag",
			Description: "Let's you display a tag",
			Options: []discord.ApplicationCommandOption{

				discord.ApplicationCommandOptionString{
					OptionName:   "name",
					Description:  "The name of the tag to display",
					Required:     true,
					Autocomplete: true,
				},
			},
		},
		CommandHandlers: map[string]handler.CommandHandler{
			"": tagHandler(b),
		},
		AutocompleteHandlers: map[string]handler.AutocompleteHandler{
			"": autoCompleteListTagHandler(b, false),
		},
	}
}

func tagHandler(b *butler.Butler) func(e *events.ApplicationCommandInteractionCreate) error {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		tag, err := b.DB.GetAndIncrement(*e.GuildID(), e.SlashCommandInteractionData().String("name"))
		if err == sql.ErrNoRows {
			return common.RespondErrMessage(e.Respond, "Tag not found")
		} else if err != nil {
			return common.RespondErr(e.Respond, err)
		}
		return e.CreateMessage(discord.MessageCreate{
			Content: tag.Content,
		})
	}
}
