package server

import (
	"embed"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/oauth2"
	"github.com/google/go-github/v93/github"

	"github.com/disgoorg/disgo-butler/server/database/dbtype"
	"github.com/disgoorg/disgo-butler/server/database/sqlc"
)

//go:embed templates/*
var templateFS embed.FS

type pageData struct {
	Title   string
	Message string
	Repos   []string
	Failed  bool
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, s.OAuth2.GenerateAuthorizationURL(oauth2.AuthorizationURLParams{
		RedirectURI: s.Cfg.BaseURL + "/github",
		Scopes:      []discord.OAuth2Scope{discord.OAuth2ScopeGuildsMembersRead, discord.OAuth2ScopeConnections, discord.OAuth2ScopeRoleConnectionsWrite},
	}), http.StatusTemporaryRedirect)
}

func (s *Server) handleGithub(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	code := query.Get("code")
	state := query.Get("state")
	if code == "" || state == "" {
		http.Redirect(w, r, s.Cfg.BaseURL+"/github/login", http.StatusTemporaryRedirect)
		return
	}

	session, _, err := s.OAuth2.StartSession(code, state)
	if err != nil {
		s.renderPage(w, http.StatusBadRequest, pageData{Title: "GitHub connection failed", Message: err.Error(), Failed: true})
		return
	}
	connections, err := s.OAuth2.GetConnections(session)
	if err != nil {
		s.renderPage(w, http.StatusInternalServerError, pageData{Title: "GitHub connection failed", Message: err.Error(), Failed: true})
		return
	}

	var connection *discord.Connection
	for i := range connections {
		if connections[i].Type == discord.ConnectionTypeGitHub {
			connection = &connections[i]
			break
		}
	}
	if connection == nil {
		s.renderPage(w, http.StatusBadRequest, pageData{Title: "GitHub connection failed", Message: "No GitHub connection found on this Discord account.", Failed: true})
		return
	}

	if err = s.DB.CreateContributor(r.Context(), sqlc.CreateContributorParams{
		Username:     connection.Name,
		AccessToken:  session.AccessToken,
		RefreshToken: session.RefreshToken,
		TokenType:    string(session.TokenType),
		Scopes:       dbtype.OAuthScopes(session.Scopes),
		Expiration:   session.Expiration,
	}); err != nil {
		s.renderPage(w, http.StatusInternalServerError, pageData{Title: "GitHub connection failed", Message: err.Error(), Failed: true})
		return
	}

	repos, err := s.DB.ListContributorRepos(r.Context())
	if err != nil {
		s.renderPage(w, http.StatusInternalServerError, pageData{Title: "GitHub connection failed", Message: err.Error(), Failed: true})
		return
	}
	contributorRepos := map[string][]*github.Contributor{}
	for _, repo := range repos {
		values := strings.SplitN(repo, "/", 2)
		githubContributors, _, listErr := s.GitHub.Repositories.ListContributors(r.Context(), values[0], values[1], nil)
		if listErr != nil {
			s.renderPage(w, http.StatusInternalServerError, pageData{Title: "GitHub connection failed", Message: listErr.Error(), Failed: true})
			return
		}
		contributorRepos[repo] = githubContributors
	}

	metadata, err := s.contributorMetadata(connection.Name, contributorRepos)
	if err != nil {
		s.renderPage(w, http.StatusInternalServerError, pageData{Title: "GitHub connection failed", Message: err.Error(), Failed: true})
		return
	}
	platform := "GitHub"
	if _, err = s.OAuth2.UpdateApplicationRoleConnection(session, s.Client.ApplicationID, discord.ApplicationRoleConnectionUpdate{
		PlatformName:     &platform,
		PlatformUsername: &connection.Name,
		Metadata:         &metadata,
	}); err != nil {
		s.renderPage(w, http.StatusInternalServerError, pageData{Title: "GitHub connection failed", Message: err.Error(), Failed: true})
		return
	}

	s.renderPage(w, http.StatusOK, pageData{
		Title:   "Contributions updated",
		Message: "Your contribution counts were updated for these repositories.",
		Repos:   repos,
	})
}

func (s *Server) renderPage(w http.ResponseWriter, status int, data pageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := s.pages.ExecuteTemplate(w, "page", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) contributorMetadata(username string, contributorRepos map[string][]*github.Contributor) (map[string]string, error) {
	metadata := make(map[string]string)
	for repo, contributors := range contributorRepos {
		for _, contributor := range contributors {
			if contributor.GetLogin() == username {
				metadata[strings.ReplaceAll(repo, "/", "_")+"_contributions"] = strconv.Itoa(contributor.GetContributions())
				break
			}
		}
	}
	return metadata, nil
}

func loadPages() *template.Template {
	return template.Must(template.New("github").ParseFS(templateFS, "templates/*.html"))
}
