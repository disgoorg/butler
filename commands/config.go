package commands

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo-butler/common"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/handler"
	"github.com/disgoorg/snowflake/v2"
)

func ConfigCommand(b *butler.Butler) handler.Command {
	return handler.Command{

		Create: discord.SlashCommandCreate{
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
								discord.ApplicationCommandOptionRole{
									Name:        "role",
									Description: "The role to assign if a user is a contributor.",
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
		},
		CommandHandlers: map[string]handler.CommandHandler{
			"aliases/add":              handleAliasesAdd(b),
			"aliases/remove":           handleAliasesRemove(b),
			"aliases/list":             handleAliasesList(b),
			"releases/add":             handleReleasesAdd(b),
			"releases/remove":          handleReleasesRemove(b),
			"releases/list":            handleReleasesList(b),
			"contributor-repos/add":    handleContributorReposAdd(b),
			"contributor-repos/remove": handleContributorReposRemove(b),
			"contributor-repos/list":   handleContributorReposList(b),
		},
	}
}

func handleAliasesAdd(b *butler.Butler) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		data := e.SlashCommandInteractionData()
		module := data.String("module")
		alias := data.String("alias")
		go func() {
			_, _ = b.DocClient.Search(context.TODO(), module)
		}()
		b.Config.Docs.Aliases[alias] = module
		if err := butler.SaveConfig(b.Config); err != nil {
			return common.RespondErr(e.Respond, err)
		}
		return common.Respondf(e.Respond, "Added alias `%s` for module `%s`.", alias, module)
	}
}

func handleAliasesRemove(b *butler.Butler) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		data := e.SlashCommandInteractionData()
		alias := data.String("alias")

		if _, ok := b.Config.Docs.Aliases[alias]; !ok {
			return common.RespondErrMessagef(e.Respond, "alias `%s` does not exist", alias)
		}

		delete(b.Config.Docs.Aliases, alias)
		if err := butler.SaveConfig(b.Config); err != nil {
			return common.RespondErr(e.Respond, err)
		}
		return common.Respondf(e.Respond, "Removed alias `%s`.", alias)
	}
}

func handleAliasesList(b *butler.Butler) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		var message string
		for alias, module := range b.Config.Docs.Aliases {
			message += fmt.Sprintf("•`%s` -> `%s`\n", alias, module)
		}
		return common.Respondf(e.Respond, "Aliases:\n%s", message)
	}
}

func handleReleasesAdd(b *butler.Butler) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		data := e.SlashCommandInteractionData()
		name := data.String("name")
		channelID := data.Snowflake("channel")
		pingRoleID := data.Snowflake("ping-role")

		webhook, err := b.Client.Rest().CreateWebhook(channelID, discord.WebhookCreate{Name: name})
		if err != nil {
			return common.RespondErr(e.Respond, err)
		}

		if b.Config.GithubReleases == nil {
			b.Config.GithubReleases = map[string]butler.GithubReleaseConfig{}
		}

		b.Config.GithubReleases[name] = butler.GithubReleaseConfig{
			WebhookID:    webhook.ID(),
			WebhookToken: webhook.Token,
			PingRole:     pingRoleID,
		}
		if err = butler.SaveConfig(b.Config); err != nil {
			return common.RespondErr(e.Respond, err)
		}
		return common.Respondf(e.Respond, "Added release announcement for `%s`.", name)
	}
}

func handleReleasesRemove(b *butler.Butler) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		data := e.SlashCommandInteractionData()
		name := data.String("name")

		if _, ok := b.Config.GithubReleases[name]; !ok {
			return common.RespondErrMessagef(e.Respond, "release `%s` does not exist", name)
		}

		delete(b.Config.GithubReleases, name)
		if err := butler.SaveConfig(b.Config); err != nil {
			return common.RespondErr(e.Respond, err)
		}
		return common.Respondf(e.Respond, "Removed release announcement for `%s`.", name)
	}
}

func handleReleasesList(b *butler.Butler) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		var message string
		for name := range b.Config.GithubReleases {
			message += fmt.Sprintf("•`%s`\n", name)
		}
		return common.Respondf(e.Respond, "Releases:\n%s", message)
	}
}

func handleContributorReposAdd(b *butler.Butler) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		data := e.SlashCommandInteractionData()
		name := data.String("name")
		roleID := data.Snowflake("role")

		if b.Config.ContributorRepos == nil {
			b.Config.ContributorRepos = map[string]snowflake.ID{}
		}

		b.Config.ContributorRepos[name] = roleID
		if err := butler.SaveConfig(b.Config); err != nil {
			return common.RespondErr(e.Respond, err)
		}
		return common.Respondf(e.Respond, "Added contributor repository `%s`.", name)
	}
}

func handleContributorReposRemove(b *butler.Butler) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		data := e.SlashCommandInteractionData()
		name := data.String("name")

		if _, ok := b.Config.ContributorRepos[name]; !ok {
			return common.RespondErrMessagef(e.Respond, "contributor repository `%s` does not exist", name)
		}

		delete(b.Config.ContributorRepos, name)
		if err := butler.SaveConfig(b.Config); err != nil {
			return common.RespondErr(e.Respond, err)
		}
		return common.Respondf(e.Respond, "Removed contributor repository `%s`.", name)
	}
}

func handleContributorReposList(b *butler.Butler) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		var message string
		for name, roleID := range b.Config.ContributorRepos {
			message += fmt.Sprintf("•`%s` -> %s\n", name, discord.RoleMention(roleID))
		}
		return common.Respondf(e.Respond, "Repositories:\n%s", message)
	}
}
