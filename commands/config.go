package commands

import (
	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo-butler/common"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

var ConfigCommand = butler.Command{
	Create: discord.SlashCommandCreate{
		CommandName: "config",
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
				},
			},
		},
	},
	CommandHandlers: map[string]butler.HandleFunc{
		"aliases/add":     handleAliasesAdd,
		"aliases/remove":  handleAliasesRemove,
		"releases/add":    handleReleasesAdd,
		"releases/remove": handleReleasesRemove,
	},
}

func handleAliasesAdd(b *butler.Butler, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	module := data.String("module")
	alias := data.String("alias")

	b.Config.Docs.Aliases[alias] = module
	if err := butler.SaveConfig(b.Config); err != nil {
		return common.RespondErr(e, err)
	}
	return common.Respondf(e, "Added alias `%s` for module `%s`.", alias, module)
}

func handleAliasesRemove(b *butler.Butler, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	alias := data.String("alias")

	if _, ok := b.Config.Docs.Aliases[alias]; !ok {
		return common.RespondErrMessagef(e, "alias `%s` does not exist", alias)
	}

	delete(b.Config.Docs.Aliases, alias)
	if err := butler.SaveConfig(b.Config); err != nil {
		return common.RespondErr(e, err)
	}
	return common.Respondf(e, "Removed alias `%s`.", alias)
}

func handleReleasesAdd(b *butler.Butler, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	name := data.String("name")
	channelID := data.Snowflake("channel")
	pingRoleID := data.Snowflake("ping-role")

	webhook, err := b.Client.Rest().CreateWebhook(channelID, discord.WebhookCreate{Name: name})
	if err != nil {
		return common.RespondErr(e, err)
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
		return common.RespondErr(e, err)
	}
	return common.Respondf(e, "Added release announcement for `%s`.", name)
}

func handleReleasesRemove(b *butler.Butler, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	name := data.String("name")

	if _, ok := b.Config.GithubReleases[name]; !ok {
		return common.RespondErrMessagef(e, "release `%s` does not exist", name)
	}

	delete(b.Config.GithubReleases, name)
	if err := butler.SaveConfig(b.Config); err != nil {
		return common.RespondErr(e, err)
	}
	return common.Respondf(e, "Removed release announcement for `%s`.", name)
}
