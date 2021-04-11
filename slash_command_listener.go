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
	}
}
