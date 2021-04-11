package main

import "github.com/DisgoOrg/disgo/api"

var commands = []api.Command{
	{
		Name:              "docs",
		Description:       "searches something in the docs",
		DefaultPermission: true,
		Options: []*api.CommandOption{
			{
				Name:        "package",
				Description: "the package to search in",
				Type:        api.CommandOptionTypeString,
				Required:    true,
			},
			{
				Name:        "identifier",
				Description: "the identifier to search for",
				Type:        api.CommandOptionTypeString,
				Required:    true,
			},
		},
	},
	{
		Name:              "reload-docs",
		Description:       "reloads disog source",
		DefaultPermission: false,
	},
}
