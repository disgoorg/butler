package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/hhhapz/doc"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

var docsCommand = discord.SlashCommandCreate{
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
}

func (s *Server) handleDocs(event *handler.CommandEvent) error {
	data := event.SlashCommandInteractionData()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	pkg, err := s.DocClient.Search(ctx, data.String("module"))
	if err != nil {
		return respondErr(event.Respond, err)
	}

	embed, selectMenu := getDocsEmbed(pkg, data.String("query"), false, false, false, false)
	return event.CreateMessage(discord.NewMessageCreate().
		WithEmbeds(embed).
		AddActionRow(selectMenu),
	)
}

func (s *Server) handleDocsAutocomplete(event *handler.AutocompleteEvent) error {
	moduleOption, moduleOptionOK := event.Data.Option("module")
	if moduleOptionOK && moduleOption.Focused {
		return s.handleModuleAutocomplete(event, event.Data.String("module"))
	}
	if option, ok := event.Data.Option("query"); ok && option.Focused {
		return s.handleQueryAutocomplete(event, event.Data.String("module"), event.Data.String("query"))
	}
	return event.AutocompleteResult(nil)
}

func (s *Server) handleModuleAutocomplete(event *handler.AutocompleteEvent, module string) error {
	choices := make([]discord.AutocompleteChoiceString, 0, 25)
	if module == "" {
		s.DocClient.WithCache(func(cache map[string]*doc.CachedPackage) {
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
		_, _ = s.DocClient.Search(ctx, module)
		s.DocClient.WithCache(func(cache map[string]*doc.CachedPackage) {
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
	return event.AutocompleteResult(s.replaceAliases(choices))
}

func (s *Server) handleQueryAutocomplete(event *handler.AutocompleteEvent, module string, query string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	pkg, err := s.DocClient.Search(ctx, module)
	var status doc.InvalidStatusError
	if errors.As(err, &status) && status == 404 {
		return event.AutocompleteResult([]discord.AutocompleteChoice{
			discord.AutocompleteChoiceString{Name: "module not found", Value: ""},
		})
	} else if err != nil {
		return event.AutocompleteResult(nil)
	}
	choices := make([]discord.AutocompleteChoiceString, 0, 25)
	if query == "" {
		choices = append(choices, discord.AutocompleteChoiceString{Name: "<Pkg Info>", Value: pkgInfo})
	}
	var symbols []string
	for _, t := range pkg.Types {
		symbols = append(symbols, t.Name)
		for _, method := range t.Methods {
			symbols = append(symbols, fmt.Sprintf("%s.%s", t.Name, method.Name))
		}
	}
	for _, function := range pkg.Functions {
		symbols = append(symbols, function.Name)
	}
	ranks := fuzzy.RankFindFold(query, symbols)
	sort.Sort(ranks)

	for _, rank := range ranks {
		if len(choices) == 25 {
			break
		}
		choices = append(choices, discord.AutocompleteChoiceString{Name: rank.Target, Value: rank.Target})
	}
	return event.AutocompleteResult(s.replaceAliases(choices))
}

func (s *Server) replaceAliases(choices []discord.AutocompleteChoiceString) []discord.AutocompleteChoice {
	s.aliasesMu.Lock()
	defer s.aliasesMu.Unlock()

	replaced := make([]discord.AutocompleteChoice, len(choices))
	for i, choice := range choices {
		for alias, module := range s.aliases {
			if strings.HasPrefix(choice.Value, module) {
				choice.Name = strings.Replace(choice.Name, module, alias, 1)
			}
		}
		replaced[i] = choice
	}
	return replaced
}

func (s *Server) warmupAliases() {
	s.aliasesMu.Lock()
	modules := make([]string, 0, len(s.aliases))
	for _, module := range s.aliases {
		modules = append(modules, module)
	}
	s.aliasesMu.Unlock()

	slog.Info("loading go module aliases")
	for _, module := range modules {
		slog.Info("loading alias", slog.String("module", module))
		if _, err := s.DocClient.Search(context.Background(), module); err != nil {
			slog.Error("failed to load module alias", slog.String("module", module), slog.Any("err", err))
		}
	}
}
