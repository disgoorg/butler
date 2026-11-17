package server

import (
	"context"
	"fmt"
	"slices"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"github.com/disgoorg/disgo-butler/server/database"
	"github.com/disgoorg/disgo-butler/server/database/sqlc"
)

var configCommand = discord.SlashCommandCreate{
	Name:        "config",
	Description: "Used to configure aliases and release announcements.",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionSubCommandGroup{
			Name:        "aliases",
			Description: "Used to configure module aliases.",
			Options: []discord.ApplicationCommandOptionSubCommand{
				{
					Name:        "add",
					Description: "Used to add a module alias.",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:        "module",
							Description: "The module you want to add an alias for.",
							Required:    true,
						},
						discord.ApplicationCommandOptionString{
							Name:        "alias",
							Description: "The alias you want to add for the module.",
							Required:    true,
						},
					},
				},
				{
					Name:        "remove",
					Description: "Used to remove a module alias.",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:        "alias",
							Description: "The alias you want to add for the module.",
							Required:    true,
						},
					},
				},
				{
					Name:        "list",
					Description: "Used to list all module aliases.",
				},
			},
		},
		discord.ApplicationCommandOptionSubCommandGroup{
			Name:        "releases",
			Description: "Used to configure release announcements.",
			Options: []discord.ApplicationCommandOptionSubCommand{
				{
					Name:        "add",
					Description: "Used to add a release announcement.",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:        "name",
							Description: "The name of the release announcement.",
							Required:    true,
						},
						discord.ApplicationCommandOptionChannel{
							Name:        "channel",
							Description: "The channel to release the announcement in.",
							Required:    true,
						},
						discord.ApplicationCommandOptionRole{
							Name:        "ping-role",
							Description: "The role you want to ping when a new release is available.",
							Required:    true,
						},
					},
				},
				{
					Name:        "remove",
					Description: "Used to remove a release announcement.",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:        "name",
							Description: "The release announcement you want to remove.",
							Required:    true,
						},
					},
				},
				{
					Name:        "list",
					Description: "Used to list all release announcements.",
				},
			},
		},
		discord.ApplicationCommandOptionSubCommandGroup{
			Name:        "contributor-repos",
			Description: "Used to configure contributor repositories.",
			Options: []discord.ApplicationCommandOptionSubCommand{
				{
					Name:        "add",
					Description: "Used to add a contributor repositories.",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:        "name",
							Description: "The name of the contributor repository.",
							Required:    true,
						},
					},
				},
				{
					Name:        "remove",
					Description: "Used to remove a contributor repositories.",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionString{
							Name:        "name",
							Description: "The contributor repository you want to remove.",
							Required:    true,
						},
					},
				},
				{
					Name:        "list",
					Description: "Used to list all contributor repositories.",
				},
			},
		},
	},
}

func (s *Server) handleAliasesAdd(event *handler.CommandEvent) error {
	data := event.SlashCommandInteractionData()
	module := data.String("module")
	alias := data.String("alias")
	go func() {
		_, _ = s.DocClient.Search(context.Background(), module)
	}()

	if err := s.DB.CreateAlias(event.Ctx, sqlc.CreateAliasParams{Alias: alias, Module: module}); err != nil {
		return respondErr(event.Respond, err)
	}
	s.aliasesMu.Lock()
	s.aliases[alias] = module
	s.aliasesMu.Unlock()
	return respondf(event.Respond, "Added alias `%s` for module `%s`.", alias, module)
}

func (s *Server) handleAliasesRemove(event *handler.CommandEvent) error {
	alias := event.SlashCommandInteractionData().String("alias")

	if _, err := s.DB.GetAlias(event.Ctx, alias); database.IsNotFound(err) {
		return respondErrMessagef(event.Respond, "alias `%s` does not exist", alias)
	} else if err != nil {
		return respondErr(event.Respond, err)
	}

	if err := s.DB.DeleteAlias(event.Ctx, alias); err != nil {
		return respondErr(event.Respond, err)
	}
	s.aliasesMu.Lock()
	delete(s.aliases, alias)
	s.aliasesMu.Unlock()
	return respondf(event.Respond, "Removed alias `%s`.", alias)
}

func (s *Server) handleAliasesList(event *handler.CommandEvent) error {
	aliases, err := s.DB.ListAliases(event.Ctx)
	if err != nil {
		return respondErr(event.Respond, err)
	}
	var message string
	for _, alias := range aliases {
		message += fmt.Sprintf("•`%s` -> `%s`\n", alias.Alias, alias.Module)
	}
	return respondf(event.Respond, "Aliases:\n%s", message)
}

func (s *Server) handleReleasesAdd(event *handler.CommandEvent) error {
	data := event.SlashCommandInteractionData()
	name := data.String("name")
	channelID := data.Snowflake("channel")
	pingRoleID := data.Snowflake("ping-role")

	created, err := s.Client.Rest.CreateWebhook(channelID, discord.WebhookCreate{Name: name})
	if err != nil {
		return respondErr(event.Respond, err)
	}

	if err = s.DB.CreateGithubRelease(event.Ctx, sqlc.CreateGithubReleaseParams{
		Name:         name,
		WebhookID:    created.ID(),
		PingRole:     pingRoleID,
		WebhookToken: created.Token,
	}); err != nil {
		return respondErr(event.Respond, err)
	}
	s.webhooksMu.Lock()
	delete(s.Webhooks, name)
	s.webhooksMu.Unlock()
	return respondf(event.Respond, "Added release announcement for `%s`.", name)
}

func (s *Server) handleReleasesRemove(event *handler.CommandEvent) error {
	name := event.SlashCommandInteractionData().String("name")
	if _, err := s.DB.GetGithubRelease(event.Ctx, name); database.IsNotFound(err) {
		return respondErrMessagef(event.Respond, "release `%s` does not exist", name)
	} else if err != nil {
		return respondErr(event.Respond, err)
	}

	if err := s.DB.DeleteGithubRelease(event.Ctx, name); err != nil {
		return respondErr(event.Respond, err)
	}
	s.webhooksMu.Lock()
	delete(s.Webhooks, name)
	s.webhooksMu.Unlock()
	return respondf(event.Respond, "Removed release announcement for `%s`.", name)
}

func (s *Server) handleReleasesList(event *handler.CommandEvent) error {
	releases, err := s.DB.ListGithubReleases(event.Ctx)
	if err != nil {
		return respondErr(event.Respond, err)
	}
	var message string
	for _, release := range releases {
		message += fmt.Sprintf("•`%s`\n", release.Name)
	}
	return respondf(event.Respond, "Releases:\n%s", message)
}

func (s *Server) handleContributorReposAdd(event *handler.CommandEvent) error {
	name := event.SlashCommandInteractionData().String("name")
	if err := s.DB.CreateContributorRepo(event.Ctx, name); err != nil {
		return respondErr(event.Respond, err)
	}
	return respondf(event.Respond, "Added contributor repository `%s`.", name)
}

func (s *Server) handleContributorReposRemove(event *handler.CommandEvent) error {
	name := event.SlashCommandInteractionData().String("name")
	repos, err := s.DB.ListContributorRepos(event.Ctx)
	if err != nil {
		return respondErr(event.Respond, err)
	}
	if !slices.Contains(repos, name) {
		return respondErrMessagef(event.Respond, "contributor repository `%s` does not exist", name)
	}
	if err = s.DB.DeleteContributorRepo(event.Ctx, name); err != nil {
		return respondErr(event.Respond, err)
	}
	return respondf(event.Respond, "Removed contributor repository `%s`.", name)
}

func (s *Server) handleContributorReposList(event *handler.CommandEvent) error {
	repos, err := s.DB.ListContributorRepos(event.Ctx)
	if err != nil {
		return respondErr(event.Respond, err)
	}
	var message string
	for _, name := range repos {
		message += fmt.Sprintf("•`%s`\n", name)
	}
	return respondf(event.Respond, "Repositories:\n%s", message)
}
