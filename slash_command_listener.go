package main

import (
	"go/ast"

	"github.com/DisgoOrg/disgo/api"
	"github.com/DisgoOrg/disgo/api/events"
)

const red = 16711680
const green = 65280

func slashCommandListener(event *events.SlashCommandEvent) {
	switch event.CommandName {
	case "docs":
		result := findInPackages(event.Option("package").String(), event.Option("identifier").String())

		var embed api.Embed
		if result == nil {
			embed = api.NewEmbedBuilder().
				SetColor(red).
				SetDescriptionf("no result for: `%s#%s` found", event.Option("package"), event.Option("identifier")).
				Build()
		} else {
			title := result.PackagePath + " - " + result.Kind.String() + " " + result.IdentifierName
			if result.Kind == ast.Fun {
				title += "("
				if result.Params != nil {
					title += *result.Params
				}
				title += ")"
				if result.Results != nil {
					title += " (" + *result.Results + ")"
				}
			}
			embed = api.NewEmbedBuilder().
				SetColor(green).
				SetTitle(title).
				SetURL("https://pkg.go.dev/github.com/DisgoOrg/" + result.PackagePath + "#" + result.PackageName).
				SetDescription(result.Comment).
				Build()
		}
		_ = event.Reply(api.NewInteractionResponseBuilder().SetEmbeds(embed).Build())

	case "reload-docs":
		go func() {
			_ = event.Acknowledge()
			var embed api.Embed

			if err := downloadDisgo(event.Disgo().RestClient()); err != nil {
				logger.Errorf("error while downloading latest source from github: %s", err)
				embed = api.NewEmbedBuilder().
					SetColor(red).
					SetDescriptionf("❌ failed to download latest disgo: %s", err).
					Build()
			} else if err := loadPackages(); err != nil {
				logger.Errorf("error while loading disgo packages: %s", err)
				embed = api.NewEmbedBuilder().
					SetColor(red).
					SetDescriptionf("❌ failed to load latest disgo packages: %s", err).
					Build()
			} else {
				embed = api.NewEmbedBuilder().
					SetColor(red).
					SetDescription("✅ loaded latest disgo").
					Build()
			}

			_, _ = event.EditOriginal(api.NewFollowupMessageBuilder().
				SetEmbeds(embed).
				Build(),
			)
		}()

	}
}
