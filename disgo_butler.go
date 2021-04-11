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

func main() {
	logger.SetLevel(logrus.InfoLevel)
	logger.Info("starting disgo-butler...")

	err := loadPackages("../disgo")
	if err != nil {
		logger.Errorf("error while loading disgo source: %s", err)
	}

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

	_, err = dgo.SetCommands(commands...)
	if err != nil {
		logger.Errorf("error registering commands: %s", err)
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
