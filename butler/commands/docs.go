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

const (
	embedDescriptionFormat = "```go\n%s```\n\n%s"
	embedTitleFormat       = "%s: %s"
	embedColor             = 0x5865f2
	embedPackageURLFormat  = "https://pkg.go.dev/%s"
	embedURLFormat         = embedPackageURLFormat + "#%s"
	typeFuncFormat         = "%s.%s"
)

func handleDocs(b *butler.Butler, e *events.ApplicationCommandInteractionEvent) error {
	println("handleDocs")
	data := e.SlashCommandInteractionData()

	pkg, err := b.DocClient.Search(context.Background(), data.String("module"))
	if err != nil {
		return common.RespondErr(e, err)
	}
	query, ok := data.OptString("query")
	var (
		embed discord.Embed
		more  bool
	)

	if !ok {
		embed, more = embedFromPackage(pkg)
	} else {
		values := strings.Split(query, ".")
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

	return e.CreateMessage(discord.NewMessageCreateBuilder().
		SetEmbeds(embed).
		AddActionRow(discord.NewSelectMenu(discord.CustomID("docs_action:"+pkg.URL), "action", options...)).
		Build(),
	)
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
	if markdown == "" {
		markdown = "No comments found."
	}
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

type UniqueChoices struct {
	set     map[string]struct{}
	choices []discord.AutocompleteChoice
}

func (c *UniqueChoices) Len() int {
	return len(c.choices)
}

func (c *UniqueChoices) AddString(name string, value string) {
	if _, ok := c.set[value]; !ok {
		c.set[value] = struct{}{}
		c.choices = append(c.choices, discord.AutocompleteChoiceString{Name: name, Value: value})
	}
}

func handleDocsAutocomplete(b *butler.Butler, e *events.AutocompleteInteractionEvent) error {
	choices := &UniqueChoices{set: map[string]struct{}{}}
	if option, ok := e.Data.StringOption("module"); ok && option.Focused() {
		if option.Value == "" {
			b.DocClient.WithCache(func(cache map[string]*doc.CachedPackage) {
				if len(cache) == 0 {
					choices.AddString("No modules found", "nothing")
					return
				}
				for _, pkg := range cache {
					if choices.Len() > 24 {
						return
					}
					choices.AddString(pkg.URL, pkg.URL)
				}
			})
		} else {
			choices.AddString(option.Value, option.Value)
			_, _ = b.DocClient.Search(context.TODO(), option.Value)
			b.DocClient.WithCache(func(cache map[string]*doc.CachedPackage) {
				var packages []string
				for name, pkg := range cache {
					packages = append(packages, name)
					for _, subName := range pkg.Subpackages {
						packages = append(packages, subName)
					}
				}
				ranks := fuzzy.RankFindFold(option.Value, packages)
				sort.Sort(ranks)

				if len(ranks) > 24 {
					ranks = ranks[:24]
				}

				for _, rank := range ranks {
					if choices.Len() > 24 {
						break
					}
					choices.AddString(rank.Target, rank.Target)
				}
			})
		}
	} else if option, ok := e.Data.StringOption("query"); ok && option.Focused() {
		pkg, err := b.DocClient.Search(context.Background(), e.Data.String("module"))
		if err != nil {
			choices.AddString("module not found", "nothing")
		} else {
			if option.Value == "" {
				for _, t := range pkg.Types {
					if choices.Len() > 24 {
						break
					}
					choices.AddString(t.Name, t.Name)
				}
				for _, f := range pkg.Functions {
					if choices.Len() > 24 {
						break
					}
					choices.AddString(f.Name, f.Name)
				}
			} else {
				var types []string
				for _, t := range pkg.Types {
					types = append(types, t.Name)
					for _, m := range t.Methods {
						types = append(types, fmt.Sprintf("%s.%s", t.Name, m.Name))
					}
				}
				for _, f := range pkg.Functions {
					types = append(types, f.Name)
				}
				ranks := fuzzy.RankFindFold(option.Value, types)
				sort.Sort(ranks)

				for _, rank := range ranks {
					if choices.Len() > 24 {
						break
					}
					choices.AddString(rank.Target, rank.Target)
				}
			}
		}
	} else {
		choices.AddString("No results found", "nothing")
	}
	return e.Result(choices.choices)
}
