package main

import (
	"github.com/DisgoOrg/disgo-butler/butler"
	"github.com/DisgoOrg/log"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetLevel(log.LevelInfo)
	log.Info()

	cfg, err := butler.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	bot := butler.New(*cfg)

	bot.StartAndBlock()
}
