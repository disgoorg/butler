package components

import (
	"context"
	"strings"

	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo-butler/common"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/handler"
)

func DocsActionComponent(b *butler.Butler) handler.Component {
	return handler.Component{
		Name:    "docs_action",
		Handler: handleDocsAction(b),
	}
}

func handleDocsAction(b *butler.Butler) func(ctx *handler.ComponentContext) error {
	return func(ctx *handler.ComponentContext) error {
		action := ctx.SelectMenuInteractionData().Values[0]
		if action == "delete" {
			if ctx.Message.Interaction.User.ID != ctx.User().ID && ctx.Member().Permissions.Missing(discord.PermissionManageMessages) {
				return common.RespondErrMessage(ctx.Respond, "You don't have permission to delete this message.")
			}
			_ = ctx.DeferUpdateMessage()
			return ctx.Client().Rest().DeleteInteractionResponse(ctx.ApplicationID(), ctx.Token())
		}
		values := strings.SplitN(ctx.Message.Embeds[0].Title, ": ", 2)
		pkg, err := b.DocClient.Search(context.Background(), values[0])
		if err != nil {
			return common.RespondErrMessagef(ctx.Respond, "Error while fetching package: %s", err)
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
			return common.RespondErrMessagef(ctx.Respond, "Unknown action: %s", action)
		}

		var query string
		if len(values) > 1 {
			query = values[1]
		}
		embed, selectMenu := butler.GetDocsEmbed(pkg, query, expandSignature, expandComment, expandMethods, expandExamples)
		if ctx.Message.Interaction.User.ID != ctx.User().ID && ctx.Member().Permissions.Missing(discord.PermissionManageMessages) {
			return ctx.CreateMessage(discord.MessageCreate{Embeds: []discord.Embed{embed}, Flags: discord.MessageFlagEphemeral})
		}
		return ctx.UpdateMessage(discord.MessageUpdate{Embeds: &[]discord.Embed{embed}, Components: &[]discord.ContainerComponent{discord.NewActionRow(selectMenu)}})
	}
}
