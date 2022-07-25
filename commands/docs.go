package commands

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo-butler/common"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/handler"
	"github.com/hhhapz/doc"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

func DocsCommand(b *butler.Butler) handler.Command {
	return handler.Command{
		Create: discord.SlashCommandCreate{
			Name:        "docs",
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
		CommandHandlers: map[string]handler.CommandHandler{
			"": handleDocs(b),
		},
		AutocompleteHandlers: map[string]handler.AutocompleteHandler{
			"": handleDocsAutocomplete(b),
		},
	}
}

func handleDocs(b *butler.Butler) func(ctx *handler.CommandContext) error {
	return func(ctx *handler.CommandContext) error {
		data := ctx.SlashCommandInteractionData()

		ctxx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		pkg, err := b.DocClient.Search(ctxx, data.String("module"))
		if err != nil {
			return common.RespondErr(ctx.Respond, err)
		}

		embed, selectMenu := butler.GetDocsEmbed(pkg, data.String("query"), false, false, false, false)

		return ctx.CreateMessage(discord.NewMessageCreateBuilder().
			SetEmbeds(embed).
			AddActionRow(selectMenu).
			Build(),
		)
	}
}

func handleDocsAutocomplete(b *butler.Butler) func(ctx *handler.AutocompleteContext) error {
	return func(ctx *handler.AutocompleteContext) error {
		moduleOption, moduleOptionOk := ctx.Data.Option("module")
		if moduleOptionOk && moduleOption.Focused {
			return handleModuleAutocomplete(b, ctx, ctx.Data.String("module"))
		}
		if option, ok := ctx.Data.Option("query"); ok && option.Focused {
			return handleQueryAutocomplete(b, ctx, ctx.Data.String("module"), ctx.Data.String("query"))
		}
		return ctx.Result(nil)
	}
}

func handleModuleAutocomplete(b *butler.Butler, ctx *handler.AutocompleteContext, module string) error {
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
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, _ = b.DocClient.Search(ctx, module)
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
	return ctx.Result(replaceAliases(b, choices))
}

func handleQueryAutocomplete(b *butler.Butler, ctx *handler.AutocompleteContext, module string, query string) error {
	ctxx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	pkg, err := b.DocClient.Search(ctxx, module)
	if err == doc.InvalidStatusError(404) {
		return ctx.Result([]discord.AutocompleteChoice{
			discord.AutocompleteChoiceString{Name: "module not found", Value: ""},
		})
	} else if err != nil {
		return ctx.Result(nil)
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
	return ctx.Result(replaceAliases(b, choices))
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
