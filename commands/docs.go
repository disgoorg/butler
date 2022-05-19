package commands

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo-butler/common"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/hhhapz/doc"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

var DocsCommand = butler.Command{
	Create: discord.SlashCommandCreate{
		CommandName: "docs",
		Description: "Provides info to the provided module, type, function, etc.",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:         "module",
				Description:  "The module to lookup. Example: github.com/disgoorg/disgo/discord",
				Required:     true,
				Autocomplete: true,
			},
			discord.ApplicationCommandOptionString{
				Name:         "query",
				Description:  "The lookup query. Example: MessageCreate",
				Required:     true,
				Autocomplete: true,
			},
		},
	},
	CommandHandlers: map[string]butler.HandleFunc{
		"": handleDocs,
	},
	AutocompleteHandlers: map[string]butler.AutocompleteHandleFunc{
		"": handleDocsAutocomplete,
	},
}

func handleDocs(b *butler.Butler, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()

	pkg, err := b.DocClient.Search(context.Background(), data.String("module"))
	if err != nil {
		return common.RespondErr(e.Respond, err)
	}

	embed, selectMenu := butler.GetDocsEmbed(pkg, data.String("query"), false, false, false, false)

	return e.CreateMessage(discord.NewMessageCreateBuilder().
		SetEmbeds(embed).
		AddActionRow(selectMenu).
		Build(),
	)
}

func handleDocsAutocomplete(b *butler.Butler, e *events.AutocompleteInteractionEvent) error {
	moduleOption, moduleOptionOk := e.Data.StringOption("module")
	if moduleOptionOk && moduleOption.Focused() {
		return handleModuleAutocomplete(b, e, moduleOption.Value)
	}
	if option, ok := e.Data.StringOption("query"); ok && option.Focused() {
		return handleQueryAutocomplete(b, e, moduleOption.Value, option.Value)
	}
	return e.Result(nil)
}

func handleModuleAutocomplete(b *butler.Butler, e *events.AutocompleteInteractionEvent, module string) error {
	choices := make([]discord.AutocompleteChoiceString, 0, 25)
	if module == "" {
		b.DocClient.WithCache(func(cache map[string]*doc.CachedPackage) {
			for _, pkg := range cache {
				if len(choices) > 24 {
					return
				}
				choices = append(choices, discord.AutocompleteChoiceString{Name: pkg.URL, Value: pkg.URL})
			}
		})
	} else {
		_, _ = b.DocClient.Search(context.TODO(), module)
		b.DocClient.WithCache(func(cache map[string]*doc.CachedPackage) {
			var packages []string
			for _, pkg := range cache {
				packages = append(packages, pkg.URL)
				packages = append(packages, pkg.Subpackages...)
			}
			ranks := fuzzy.RankFindFold(module, packages)
			sort.Sort(ranks)
			for _, rank := range ranks {
				if len(choices) > 24 {
					break
				}
				choices = append(choices, discord.AutocompleteChoiceString{Name: rank.Target, Value: rank.Target})
			}
		})
	}
	return e.Result(replaceAliases(b, choices))
}

func handleQueryAutocomplete(b *butler.Butler, e *events.AutocompleteInteractionEvent, module string, query string) error {
	pkg, err := b.DocClient.Search(context.Background(), module)
	if err == doc.InvalidStatusError(404) {
		return e.Result([]discord.AutocompleteChoice{
			discord.AutocompleteChoiceString{Name: "module not found", Value: ""},
		})
	} else if err != nil {
		return e.Result(nil)
	}
	choices := make([]discord.AutocompleteChoiceString, 0, 25)
	if query == "" {
		choices = append(choices, discord.AutocompleteChoiceString{Name: "<Pkg Info>", Value: butler.PkgInfo})
	}
	var symbols []string
	for _, t := range pkg.Types {
		symbols = append(symbols, t.Name)
		for _, m := range t.Methods {
			symbols = append(symbols, fmt.Sprintf("%s.%s", t.Name, m.Name))
		}
	}
	for _, f := range pkg.Functions {
		symbols = append(symbols, f.Name)
	}
	ranks := fuzzy.RankFindFold(query, symbols)
	sort.Sort(ranks)

	for _, rank := range ranks {
		if len(choices) == 25 {
			break
		}
		choices = append(choices, discord.AutocompleteChoiceString{Name: rank.Target, Value: rank.Target})
	}
	return e.Result(replaceAliases(b, choices))
}

func replaceAliases(b *butler.Butler, choices []discord.AutocompleteChoiceString) []discord.AutocompleteChoice {
	newChoices := make([]discord.AutocompleteChoice, len(choices))
	for i, choice := range choices {
		for alias, module := range b.Config.Docs.Aliases {
			if strings.HasPrefix(choice.Value, module) {
				choice.Name = strings.Replace(choice.Name, module, alias, 1)
			}
		}
		newChoices[i] = choice
	}
	return newChoices
}
