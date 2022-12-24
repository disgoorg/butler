package routes

import (
	"context"
	"embed"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo/discord"
)

//go:embed templates/*
var templateFS embed.FS

func HandleLogin(b *butler.Butler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, b.OAuth2.GenerateAuthorizationURL(b.Config.BaseURL+"/github", discord.PermissionsNone, 0, false, discord.OAuth2ScopeGuildsMembersRead, discord.OAuth2ScopeConnections), http.StatusTemporaryRedirect)
	}
}

func HandleGithub(b *butler.Butler) http.HandlerFunc {
	t := template.Must(template.New("github").ParseFS(templateFS, "templates/*.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			query = r.URL.Query()
			code  = query.Get("code")
			state = query.Get("state")
		)
		if code == "" || state == "" {
			http.Redirect(w, r, b.Config.BaseURL+"/github/login", http.StatusTemporaryRedirect)
			return
		}

		session, err := b.OAuth2.StartSession(code, state, state)
		if err != nil {
			httpError(w, err)
			return
		}
		connections, err := b.OAuth2.GetConnections(session)
		if err != nil {
			httpError(w, err)
			return
		}

		var conn *discord.Connection
		for _, connection := range connections {
			if connection.Type == discord.ConnectionTypeGitHub {
				conn = &connection
				break
			}
		}
		if conn == nil {
			if err = t.ExecuteTemplate(w, "error.html", map[string]any{
				"Error": "No GitHub connection found",
			}); err != nil {
				httpError(w, err)
			}
			return
		}

		metadata := make(map[string]string)
		for _, repo := range b.Config.ContributorRepos {
			values := strings.SplitN(repo, "/", 2)
			contributors, _, err := b.GitHubClient.Repositories.ListContributors(context.TODO(), values[0], values[1], nil)
			if err != nil {
				httpError(w, err)
				return
			}
			for _, contributor := range contributors {
				if contributor.GetLogin() == conn.Name {
					// need at least 10 contributions
					contributions := 0
					if contributor.Contributions != nil {
						contributions = *contributor.Contributions
					}
					metadata[repo] = strconv.Itoa(contributions)
					break
				}
			}

		}

		if _, err = b.OAuth2.UpdateApplicationRoleConnection(session, b.Client.ApplicationID(), discord.ApplicationRoleConnectionUpdate{
			PlatformName: &conn.Name,
			Metadata:     &metadata,
		}); err != nil {
			httpError(w, err)
			return
		}

		if err = t.ExecuteTemplate(w, "response.html", map[string]any{
			"Repos": b.Config.ContributorRepos,
		}); err != nil {
			httpError(w, err)
		}
	}
}

func httpError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
