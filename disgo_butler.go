package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/DisgoOrg/disgo"
	"github.com/DisgoOrg/disgo/api"
	"github.com/DisgoOrg/disgo/api/endpoints"
	"github.com/DisgoOrg/disgo/api/events"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

var (
	disgoGuild = api.Snowflake(os.Getenv("guild"))
	adminRole  = api.Snowflake(os.Getenv("admin_role"))
)

func main() {
	logger.SetLevel(logrus.InfoLevel)
	logger.Info("starting disgo-butler...")

	dgo, err := disgo.NewBuilder(endpoints.Token(os.Getenv("token"))).
		SetLogger(logger).
		SetIntents(api.IntentsGuilds | api.IntentsGuildMessages).
		SetMemberCachePolicy(api.MemberCachePolicyNone).
		AddEventListeners(&events.ListenerAdapter{
			OnSlashCommand: slashCommandListener,
		}).
		Build()
	if err != nil {
		logger.Fatalf("error while building disgo instance: %s", err)
		return
	}

	go func() {
		logger.Infof("downloading latest source from github...")
		err := downloadDisgo(dgo.RestClient())
		if err != nil {
			logger.Errorf("error while downloading latest source from github: %s", err)
			return
		}
		logger.Infof("downloaded latest source from github")

		logger.Infof("loading disgo packages...")
		err = loadPackages()
		if err != nil {
			logger.Errorf("error while loading disgo packages: %s", err)
			return
		}
		logger.Infof("loaded disgo packages")
	}()

	cmds, err := dgo.SetCommands(commands...)
	if err != nil {
		logger.Errorf("error registering commands: %s", err)
	}
	for _, cmd := range cmds {
		if cmd.Name == "reload-docs" {
			_ = cmd.SetPermissions(disgoGuild, api.CommandPermission{
				ID:         adminRole,
				Type:       api.CommandPermissionTypeRole,
				Permission: true,
			})
		}
	}

	err = dgo.Connect()
	if err != nil {
		logger.Fatalf("error while connecting to discord: %s", err)
	}

	defer dgo.Close()

	logger.Infof("Bot is now running. Press CTRL-C to exit.")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-s
}
