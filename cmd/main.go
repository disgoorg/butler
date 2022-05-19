package main

import (
	"flag"

	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo-butler/commands"
	"github.com/disgoorg/disgo-butler/components"
	"github.com/disgoorg/disgo-butler/routes"
	"github.com/disgoorg/log"
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
	r.Post("/github", routes.HandleGithub(b))
	r.Get("/login", routes.HandleLogin(b))
	b.SetupRoutes(r)

	b.SetupBot()
	b.SetupDB(*shouldSyncDBTables)
	b.SetupCommands(*shouldSyncCommands,
		commands.PingCommand,
		commands.InfoCommand,
		commands.DocsCommand,
		commands.TagCommand,
		commands.TagsCommand,
		commands.ConfigCommand,
	)
	b.SetupComponents(
		components.DocsActionComponent,
	)
	b.StartAndBlock()
}
