package handlers

import "github.com/disgoorg/disgo-butler/butler"

var Handlers = map[string]butler.HTTPHandleFunc{
	"/login":  handleDiscordLogin,
	"/github": handleGithubWebhook,
}
