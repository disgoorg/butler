package server

import (
	"context"
	"regexp"
	"strings"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
	gopiston "github.com/milindmadhukar/go-piston"
)

var discordCodeblockRegex = regexp.MustCompile(`(?s)\x60\x60\x60(?P<language>\w+)\n(?P<code>.+)\x60\x60\x60`)

var evalCommand = discord.MessageCommandCreate{
	Name: "eval",
}

func (s *Server) handleEval(event *handler.CommandEvent) error {
	message := event.MessageCommandInteractionData().TargetMessage()
	return s.eval(event.Client(), event.ApplicationCommandInteraction, event.Respond, message.Content, message.ID, false)
}

func (s *Server) eval(client *bot.Client, interaction discord.Interaction, respondFunc events.InteractionResponderFunc, content string, messageID snowflake.ID, update bool) error {
	runtimes, err := s.Piston.GetRuntimes(context.Background())
	if err != nil {
		return respondErr(respondFunc, err)
	}

	matches := discordCodeblockRegex.FindStringSubmatch(content)
	if len(matches) == 0 {
		return respondErrMessage(respondFunc, "no codeblock found")
	}
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
		return respondErrMessagef(respondFunc, "Language %s is not supported", rawLanguage)
	}

	if update {
		if err = respondFunc(discord.InteractionResponseTypeUpdateMessage, discord.NewMessageUpdate().
			WithContent("Running...").
			WithEmbeds([]discord.Embed{}...).
			ClearComponents(),
		); err != nil {
			return err
		}
	} else if err = respondFunc(discord.InteractionResponseTypeDeferredCreateMessage, nil); err != nil {
		return err
	}

	result, err := s.Piston.Execute(context.Background(), language, "", []gopiston.Code{{Content: code}})
	var output discord.Embed
	if err != nil {
		output = discord.Embed{
			Title:       "Eval",
			Description: "```\n" + err.Error() + "\n```",
			Fields: []discord.EmbedField{
				{Name: "Status", Value: "Error"},
				{Name: "Duration", Value: "0s"},
			},
		}
	} else {
		output = discord.Embed{
			Title:       "Eval",
			Description: "```\n" + result.GetOutput() + "\n```",
			Fields: []discord.EmbedField{
				{Name: "Status", Value: "Success"},
				{Name: "Duration", Value: "0s"},
			},
		}
	}

	_, err = client.Rest.UpdateInteractionResponse(interaction.ApplicationID(), interaction.Token(), discord.NewMessageUpdate().
		ClearContent().
		WithEmbeds(output).
		AddActionRow(
			discord.NewPrimaryButton("", "/eval/rerun/"+messageID.String()).WithEmoji(discord.ComponentEmoji{Name: "🔁"}),
			discord.NewDangerButton("", "/eval/delete").WithEmoji(discord.ComponentEmoji{Name: "🗑️"}),
		),
	)
	return err
}
