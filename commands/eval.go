package commands

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo-butler/common"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/handler"
	"github.com/milindmadhukar/go-piston"
)

var discordCodeblockRegex = regexp.MustCompile(`(?s)\x60\x60\x60(?P<language>\w+)\n(?P<code>.+)\x60\x60\x60`)

func Eval(b *butler.Butler) handler.Command {
	return handler.Command{
		Create: discord.MessageCommandCreate{
			Name: "eval",
		},
		CommandHandlers: map[string]handler.CommandHandler{
			"": handleEval(b),
		},
	}
}

func handleEval(b *butler.Butler) handler.CommandHandler {
	return func(ctx *handler.CommandContext) error {
		runtimes, err := b.PistonClient.GetRuntimes()
		if err != nil {
			return common.RespondErr(ctx.Respond, err)
		}

		data := ctx.MessageCommandInteractionData()

		matches := discordCodeblockRegex.FindStringSubmatch(data.TargetMessage().Content)
		rawLanguage := matches[discordCodeblockRegex.SubexpIndex("language")]
		code := matches[discordCodeblockRegex.SubexpIndex("code")]

		var language string
	runtimeLoop:
		for _, runtime := range *runtimes {
			if strings.EqualFold(runtime.Language, rawLanguage) {
				language = runtime.Language
				break
			}
			for _, alias := range runtime.Aliases {
				if strings.EqualFold(alias, rawLanguage) {
					language = runtime.Language
					break runtimeLoop
				}
			}
		}
		if language == "" {
			return common.RespondErrMessagef(ctx.Respond, "Language %s is not supported", rawLanguage)
		}

		if err = ctx.DeferCreateMessage(false); err != nil {
			return err
		}

		rs, err := b.PistonClient.Execute(language, "", []gopiston.Code{{Content: code}})
		var output discord.Embed
		if err != nil {
			output = discord.Embed{
				Title:       "Eval",
				Description: err.Error(),
				Fields: []discord.EmbedField{
					{
						Name:  "Status",
						Value: "Error",
					},
					{
						Name:  "Duration",
						Value: "0s",
					},
				},
			}
		} else {
			output = discord.Embed{
				Title:       "Eval",
				Description: fmt.Sprintf("%s", rs.GetOutput()),
				Fields: []discord.EmbedField{
					{
						Name:  "Status",
						Value: "Success",
					},
					{
						Name:  "Duration",
						Value: "0s",
					},
				},
			}
		}

		_, err = ctx.Client().Rest().UpdateInteractionResponse(ctx.ApplicationID(), ctx.Token(), discord.MessageUpdate{
			Embeds: &[]discord.Embed{output},
		})
		return err
	}
}
