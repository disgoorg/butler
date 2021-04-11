package main

import (
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
			embed = api.NewEmbedBuilder().
				SetColor(green).
				SetTitle(result.PackagePath + " - " + result.Kind+" "+result.IdentifierName).
				SetURL("https://pkg.go.dev/github.com/DisgoOrg/" + result.PackagePath + "#" + result.PackageName).
				SetDescription(result.Comment).
				Build()
		}
		_ = event.Reply(api.NewInteractionResponseBuilder().SetEmbeds(embed).Build())
	}
}
