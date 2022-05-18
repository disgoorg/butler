package main

import (
	"github.com/disgoorg/disgo-butler/butler"
	commands2 "github.com/disgoorg/disgo-butler/commands"
	"github.com/disgoorg/disgo-butler/components"
	"github.com/disgoorg/log"
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
	b.SetupBot()
	b.SetupDB()
	b.SetupCommands(
		commands2.PingCommand,
		commands2.InfoCommand,
		commands2.DocsCommand,
		commands2.TagCommand,
		commands2.TagsCommand,
	)
	b.SetupComponents(
		components.ExpandComponent,
	)
	b.StartAndBlock()
}
