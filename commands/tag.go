package commands

import (
	"database/sql"

	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo-butler/common"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

var TagCommand = butler.Command{
	Create: discord.SlashCommandCreate{
		CommandName: "tag",
		Description: "let's you display a tag",
		Options: []discord.ApplicationCommandOption{

			discord.ApplicationCommandOptionString{
				Name:        "name",
				Description: "the name of the tag to display",
				Required:    true,
			},
		},
	},
	CommandHandlers: map[string]butler.HandleFunc{
		"": tagHandler,
	},
}

func tagHandler(b *butler.Butler, e *events.ApplicationCommandInteractionEvent) error {
	tag, err := b.DB.GetAndIncrement(*e.GuildID(), e.SlashCommandInteractionData().String("name"))
	if err == sql.ErrNoRows {
		return common.RespondErrMessage(e, "Tag not found")
	} else if err != nil {
		return common.RespondErr(e, err)
	}
	return e.CreateMessage(discord.MessageCreate{
		Content: tag.Content,
	})
}
