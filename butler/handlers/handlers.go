package handlers

import "github.com/DisgoOrg/disgo-butler/butler"

var Handlers = map[string]butler.HTTPHandleFunc{
	"/login":  handleDiscordLogin,
	"/github": handleGithubWebhook,
}
