package main

import (
	"flag"

	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo-butler/commands"
	"github.com/disgoorg/disgo-butler/components"
	"github.com/disgoorg/disgo-butler/routes"
	"github.com/disgoorg/disgo/handler"
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

	cr := handler.New()
	cr.Route("/config", func(cr handler.Router) {
		cr.Route("/aliases", func(cr handler.Router) {
			cr.HandleCommand("/add", commands.HandleAliasesAdd(b))
			cr.HandleCommand("/remove", commands.HandleAliasesRemove(b))
			cr.HandleCommand("/list", commands.HandleAliasesList(b))
		})
		cr.Route("/releases", func(cr handler.Router) {
			cr.HandleCommand("/add", commands.HandleReleasesAdd(b))
			cr.HandleCommand("/remove", commands.HandleReleasesRemove(b))
			cr.HandleCommand("/list", commands.HandleReleasesList(b))
		})
		cr.Route("/contributor-repos", func(cr handler.Router) {
			cr.HandleCommand("/add", commands.HandleContributorReposAdd(b))
			cr.HandleCommand("/remove", commands.HandleContributorReposRemove(b))
			cr.HandleCommand("/list", commands.HandleContributorReposList(b))
		})
	})
	cr.Route("/docs", func(cr handler.Router) {
		cr.HandleCommand("/", commands.HandleDocs(b))
		cr.HandleAutocomplete("/", commands.HandleDocsAutocomplete(b))
	})
	cr.HandleComponent("docs_action", components.HandleDocsAction(b))
	cr.HandleCommand("/eval", commands.HandleEval(b))
	cr.HandleCommand("/info", commands.HandleInfo(b))
	cr.HandleCommand("/ping", commands.HandlePing)
	cr.Route("/tag", func(cr handler.Router) {
		cr.HandleCommand("/", commands.HandleTag(b))
		cr.HandleAutocomplete("/", commands.HandleTagListAutoComplete(b, false))
	})
	cr.Route("/tags", func(cr handler.Router) {
		cr.HandleCommand("/create", commands.HandleCreateTag(b))
		cr.HandleCommand("/edit", commands.HandleEditTag(b))
		cr.HandleCommand("/delete", commands.HandleDeleteTag(b))
		cr.HandleCommand("/info", commands.HandleTagInfo(b))
		cr.HandleCommand("/list", commands.HandleListTags(b))
		cr.HandleAutocomplete("/edit", commands.HandleTagListAutoComplete(b, true))
		cr.HandleAutocomplete("/delete", commands.HandleTagListAutoComplete(b, true))
		cr.HandleAutocomplete("/info", commands.HandleTagListAutoComplete(b, false))
		cr.HandleAutocomplete("/list", commands.HandleTagListAutoComplete(b, false))
	})
	cr.HandleCommand("/close-ticket", commands.HandleCloseTicket(b))
	b.SetupBot(cr)
	b.SetupDB(*shouldSyncDBTables)

	if *shouldSyncCommands {
		var guildIDs []snowflake.ID
		if cfg.DevMode {
			guildIDs = append(guildIDs, cfg.GuildID)
		}
		b.SyncCommands(commands.Commands, guildIDs...)
	}
	b.StartAndBlock()
}
