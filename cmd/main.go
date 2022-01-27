package main

import (
	"github.com/DisgoOrg/disgo-butler/butler"
	"github.com/DisgoOrg/disgo-butler/butler/commands"
	"github.com/DisgoOrg/disgo-butler/butler/components"
	"github.com/DisgoOrg/disgo-butler/butler/handlers"
	"github.com/DisgoOrg/log"
)

func main() {
	cfg, err := butler.LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetLevel(cfg.LogLevel)
	log.Info("starting Disgo-Butler...")

	b := butler.New(*cfg)
	b.SetupHTTPHandlers(handlers.Handlers)
	b.SetupBot()
	b.SetupCommands(commands.Commands)
	b.SetupComponents(components.Components)
	b.StartAndBlock()
}
