package commands

import "github.com/disgoorg/disgo/discord"

var Commands = []discord.ApplicationCommandCreate{
	configCommand,
	docsCommand,
	evalCommand,
	infoCommand,
	pingCommand,
	tagCommand,
	tagsCommand,
	ticketCommand,
}
