package commands

import (
	"database/sql"

	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo-butler/common"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/handler"
)

func TagCommand(b *butler.Butler) handler.Command {
	return handler.Command{
		Create: discord.SlashCommandCreate{
			Name:        "tag",
			Description: "Let's you display a tag",
			Options: []discord.ApplicationCommandOption{

				discord.ApplicationCommandOptionString{
					Name:         "name",
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

func tagHandler(b *butler.Butler) func(ctx *handler.CommandContext) error {
	return func(ctx *handler.CommandContext) error {
		tag, err := b.DB.GetAndIncrement(*ctx.GuildID(), ctx.SlashCommandInteractionData().String("name"))
		if err == sql.ErrNoRows {
			return common.RespondErrMessage(ctx.Respond, "Tag not found")
		} else if err != nil {
			return common.RespondErr(ctx.Respond, err)
		}
		return ctx.CreateMessage(discord.MessageCreate{
			Content: tag.Content,
		})
	}
}
