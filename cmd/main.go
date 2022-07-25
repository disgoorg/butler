package main

import (
	"flag"

	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo-butler/commands"
	"github.com/disgoorg/disgo-butler/components"
	"github.com/disgoorg/disgo-butler/routes"
	"github.com/disgoorg/log"
	"github.com/disgoorg/snowflake/v2"
	"github.com/go-chi/chi/v5"
)

const version = "development"

var (
	shouldSyncDBTables *bool
	shouldSyncCommands *bool
)

func init() {
	shouldSyncDBTables = flag.Bool("sync-db", false, "Whether to sync the database tables")
	shouldSyncCommands = flag.Bool("sync-commands", false, "Whether to sync the commands")
	flag.Parse()
}

func main() {
	cfg, err := butler.LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	logger := log.New(log.LstdFlags | log.Lshortfile)
	logger.SetLevel(cfg.LogLevel)
	logger.Info("starting Disgo-Butler...")

	b := butler.New(logger, version, *cfg)

	r := chi.NewRouter()
	r.Route("/github", func(r chi.Router) {
		r.Get("/", routes.HandleGithub(b))
		r.Get("/login", routes.HandleLogin(b))
		r.Post("/webhook", routes.HandleGithubWebhook(b))
	})
	b.SetupRoutes(r)

	b.SetupBot()
	b.SetupDB(*shouldSyncDBTables)
	b.Handler.AddCommands(
		commands.PingCommand,
		commands.InfoCommand(b),
		commands.DocsCommand(b),
		commands.TagCommand(b),
		commands.TagsCommand(b),
		commands.ConfigCommand(b),
		commands.TicketCommand(b),
	)
	b.Handler.AddComponents(components.DocsActionComponent(b))
	if *shouldSyncCommands {
		var guildIDs []snowflake.ID
		if cfg.DevMode {
			guildIDs = append(guildIDs, cfg.GuildID)
		}
		b.Handler.SyncCommands(b.Client, guildIDs...)
	}
	b.StartAndBlock()
}
