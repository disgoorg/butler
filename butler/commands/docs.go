package commands

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/DisgoOrg/disgo-butler/butler"
	"github.com/DisgoOrg/disgo-butler/common"
	"github.com/DisgoOrg/disgo/core/events"
	"github.com/DisgoOrg/disgo/discord"
	"github.com/DisgoOrg/disgo/rest"
	"github.com/DisgoOrg/log"
	"github.com/hhhapz/doc"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

const (
	embedDescriptionFormat = "```go\n%s```\n\n%s"
	embedTitleFormat       = "%s: %s"
	embedColor             = 0x5865f2
	embedPackageURLFormat  = "https://pkg.go.dev/%s"
	embedURLFormat         = embedPackageURLFormat + "#%s"
)

func newCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 3*time.Second)
}

func handleDocs(b *butler.Butler, e *events.ApplicationCommandInteractionEvent) {
	println("handleDocs")
	data := e.SlashCommandInteractionData()
	ctx, cancel := newCtx()
	defer cancel()

	pkg, err := b.DocClient.Search(ctx, *data.Options.String("module"))
	if err != nil {
		common.RespondError(e, err)
		return
	}
	query := data.Options.StringOption("query")
	var (
		embed discord.Embed
		more  bool
	)

	if query == nil {
		embed, more = embedFromPackage(pkg)
	} else {
		values := strings.Split(query.Value, ".")
		for i := range values {
			values[i] = strings.ToLower(values[i])
		}
		if tp, ok := pkg.Types[values[0]]; ok {
			if len(values) > 1 {
				if mt, ok := tp.Methods[values[1]]; ok {
					embed, more = embedFromMethod(pkg, mt)
				} else if fn, ok := tp.TypeFunctions[values[1]]; ok {
					embed, more = embedFromFunc(pkg, fn)
				}
			} else {
				embed, more = embedFromType(pkg, tp)
			}
		} else if fn, ok := pkg.Functions[values[0]]; ok {
			embed, more = embedFromFunc(pkg, fn)
		}
	}

	options := []discord.SelectMenuOption{
		discord.NewSelectMenuOption("delete", "delete").WithEmoji(discord.ComponentEmoji{Name: "❌"}),
	}
	if more {
		options = append(options, discord.NewSelectMenuOption("expand", "expand").WithEmoji(discord.ComponentEmoji{Name: "🔼"}))
	}

	messageBuilder := discord.NewMessageCreateBuilder().
		SetEmbeds(embed).AddActionRow(discord.NewSelectMenu(discord.CustomID("docs_action:"+pkg.URL), "action", options...))

	if err = e.Create(messageBuilder.Build()); err != nil {
		if rErr, ok := err.(*rest.Error); ok {
			log.Errorf("Error sending message: %s", rErr.RsBody)
			return
		}
		log.Errorf("Error creating message: %s", err)
	}
}

func embedFromPackage(pkg doc.Package) (discord.Embed, bool) {
	description := pkg.Overview.Markdown()
	var more bool
	if len(description) > 1024 {
		description = description[:1023] + "…"
		more = true
	}
	return discord.Embed{
		Title:       pkg.URL,
		URL:         fmt.Sprintf(embedPackageURLFormat, pkg.URL),
		Description: description,
		Color:       embedColor,
	}, more
}

func embedFromMethod(pkg doc.Package, m doc.Method) (discord.Embed, bool) {
	description, more := formatDescription(m.Signature, m.Comment)
	return discord.Embed{
		Title:       fmt.Sprintf(embedTitleFormat, pkg.URL, m.Name),
		URL:         fmt.Sprintf(embedURLFormat, pkg.URL, m.Name),
		Description: description,
		Color:       embedColor,
	}, more
}

func embedFromFunc(pkg doc.Package, f doc.Function) (discord.Embed, bool) {
	description, more := formatDescription(f.Signature, f.Comment)
	return discord.Embed{
		Title:       fmt.Sprintf(embedTitleFormat, pkg.URL, f.Name),
		URL:         fmt.Sprintf(embedURLFormat, pkg.URL, f.Name),
		Description: description,
		Color:       embedColor,
	}, more
}

func embedFromType(pkg doc.Package, t doc.Type) (discord.Embed, bool) {
	description, more := formatDescription(t.Signature, t.Comment)
	return discord.Embed{
		Title:       fmt.Sprintf(embedTitleFormat, pkg.URL, t.Name),
		URL:         fmt.Sprintf(embedURLFormat, pkg.URL, t.Name),
		Description: description,
		Color:       embedColor,
	}, more
}

func formatDescription(signature string, comment doc.Comment) (string, bool) {
	var more bool
	lines := strings.Split(signature, "\n")
	if len(lines) > 6 {
		more = true
		signature = lines[0] + "\n..."
	}

	markdown := comment.Markdown()
	lines = strings.Split(markdown, "\n")
	if len(lines) > 6 {
		more = true
		markdown = lines[0] + "\n..."
	}
	if len(markdown) > 1024 {
		more = true
		markdown = markdown[:1023] + "…"
	}
	return fmt.Sprintf(embedDescriptionFormat, signature, markdown), more
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
