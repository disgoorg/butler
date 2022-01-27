package commands

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/DisgoOrg/disgo-butler/butler"
	"github.com/DisgoOrg/disgo-butler/common"
	"github.com/DisgoOrg/disgo/core/events"
	"github.com/DisgoOrg/disgo/discord"
	"github.com/hhhapz/doc"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

func newCtx() (context.Context, context.CancelFunc) {
	ctx := context.Background()
	return context.WithTimeout(ctx, 3*time.Second)
}

func handleDocs(b *butler.Butler, e *events.ApplicationCommandInteractionEvent) {
	ctx, cancel := newCtx()
	data := strings.Split(*e.SlashCommandInteractionData().Options.String("module"), "/")
	var (
		module string
		subPackages []string
	)
	module = data[0]+data[1]+data[2]
	if len(data) > 3 {
		subPackages = data[3:]
	}

	pkg, err := b.DocClient.Search(ctx, module)
	if err != nil {
		common.RespondError(e, err)
		return
	}
	if len(subPackages) > 0 {
		pkg.Subpackages
	}

	cancel()
}

func handleDocsAutocomplete(b *butler.Butler, e *events.AutocompleteInteractionEvent) {
	opts := make([]discord.AutocompleteChoice, 25)
	if option := e.Data.Options.StringOption("module"); option != nil && option.Focused() {
		b.DocClient.WithCache(func(cache map[string]*doc.CachedPackage) {
			packages := make([]string, len(cache))
			i := 0
			for name := range cache {
				packages[i] = name
				i++
			}
			ranks := fuzzy.RankFind(option.Value, packages)
			sort.Sort(ranks)

			opts = append(opts, discord.AutocompleteChoiceString{
				Name:  option.Value,
				Value: option.Value,
			})
			for _, rank := range ranks {
				if len(opts) >= 25 {
					break
				}
				opts = append(opts, discord.AutocompleteChoiceString{
					Name:  rank.Target,
					Value: rank.Target,
				})
			}
		})
	} else if option = e.Data.Options.StringOption("type"); option != nil && option.Focused() {
		//ctx, cancel := newCtx()
		//pkg, err := b.DocClient.Search(ctx, "")
		//cancel()
	}
}
