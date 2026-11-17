package server

import "github.com/disgoorg/disgo/discord"

var commands = []discord.ApplicationCommandCreate{
	configCommand,
	docsCommand,
	evalCommand,
	infoCommand,
	pingCommand,
	tagCommand,
	tagsCommand,
	ticketCommand,
}
