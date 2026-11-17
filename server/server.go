package server

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgo/httpserver"
	"github.com/disgoorg/disgo/oauth2"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/disgo/webhook"
	"github.com/disgoorg/paginator"
	"github.com/disgoorg/snowflake/v2"
	"github.com/google/go-github/v93/github"
	"github.com/hhhapz/doc"
	"github.com/hhhapz/doc/pkgsite"
	gopiston "github.com/milindmadhukar/go-piston"

	"github.com/disgoorg/disgo-butler/server/database"
	"github.com/disgoorg/disgo-butler/server/database/dbtype"
	"github.com/disgoorg/disgo-butler/server/database/sqlc"
)

type Server struct {
	Cfg       Config
	Version   string
	DB        *database.Database
	Client    *bot.Client
	OAuth2    *oauth2.Client
	Paginator *paginator.Manager
	DocClient doc.CachedSearcher
	GitHub    *github.Client
	Piston    *gopiston.Client
	ModMail   *ModMail
	pages     *template.Template

	aliases   map[string]string
	aliasesMu sync.Mutex

	Webhooks   map[string]*webhook.Client
	webhooksMu sync.Mutex

	contributorCancel context.CancelFunc
}

type templatePages = interface {
	ExecuteTemplate(wr interface{ Write([]byte) (int, error) }, name string, data any) error
}

func New(ctx context.Context, cfg Config, version string) (*Server, error) {
	db, err := database.New(ctx, cfg.Database)
	if err != nil {
		return nil, err
	}

	aliasRows, err := db.ListAliases(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}
	aliases := make(map[string]string, len(aliasRows))
	for _, alias := range aliasRows {
		aliases[alias.Alias] = alias.Module
	}

	threadRows, err := db.ListModMailThreads(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}
	threads := make(map[snowflake.ID]snowflake.ID, len(threadRows))
	for _, thread := range threadRows {
		threads[thread.ChannelID] = thread.ThreadID
	}

	mux := http.NewServeMux()
	router := handler.New()
	server := &Server{
		Cfg:       cfg,
		Version:   version,
		DB:        db,
		Paginator: paginator.New(),
		Piston:    gopiston.CreateDefaultClient(),
		aliases:   aliases,
		Webhooks:  map[string]*webhook.Client{},
		pages:     loadPages(),
	}
	server.ModMail = newModMail(cfg.ModMail, db, threads)
	server.routes(mux, router)

	client, err := disgo.New(cfg.Token,
		bot.WithLogger(slog.Default()),
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(gateway.IntentGuildMessages|gateway.IntentDirectMessages|gateway.IntentGuildMessageTyping|gateway.IntentDirectMessageTyping|gateway.IntentMessageContent),
			gateway.WithCompression(gateway.CompressionZlibStream),
			gateway.WithPresenceOpts(
				gateway.WithPlayingActivity("loading..."),
				gateway.WithOnlineStatus(discord.OnlineStatusDND),
			),
		),
		bot.WithCacheConfigOpts(cache.WithCaches(cache.FlagGuilds)),
		bot.WithEventListenerFunc(server.onReady),
		bot.WithEventListeners(router, server.Paginator, server.ModMail),
		bot.WithHTTPServerConfigOpts(cfg.Interactions.PublicKey,
			httpserver.WithServeMux(mux),
			httpserver.WithAddress(cfg.Interactions.Address),
			httpserver.WithURL(cfg.Interactions.URL),
		),
	)
	if err != nil {
		db.Close()
		return nil, err
	}
	server.Client = client
	server.OAuth2 = oauth2.New(client.ApplicationID, cfg.Secret)
	gitHubClient, err := github.NewClient(github.WithHTTPClient(client.Rest.HTTPClient()))
	if err != nil {
		client.Close(context.Background())
		db.Close()
		return nil, err
	}
	server.GitHub = gitHubClient
	server.DocClient = doc.NewCachedSearcher(pkgsite.Parser, doc.WithClient(client.Rest.HTTPClient()))

	go server.warmupAliases()
	if err = server.registerLinkedRoles(ctx); err != nil {
		slog.Error("failed to register linked roles", slog.Any("err", err))
	}
	return server, nil
}

func (s *Server) routes(mux *http.ServeMux, router handler.Router) {
	mux.HandleFunc("GET /github", s.handleGithub)
	mux.HandleFunc("GET /github/login", s.handleLogin)
	mux.HandleFunc("POST /github/webhook", s.handleGithubWebhook)

	router.Route("/config", func(router handler.Router) {
		router.Route("/aliases", func(router handler.Router) {
			router.Command("/add", s.handleAliasesAdd)
			router.Command("/remove", s.handleAliasesRemove)
			router.Command("/list", s.handleAliasesList)
		})
		router.Route("/releases", func(router handler.Router) {
			router.Command("/add", s.handleReleasesAdd)
			router.Command("/remove", s.handleReleasesRemove)
			router.Command("/list", s.handleReleasesList)
		})
		router.Route("/contributor-repos", func(router handler.Router) {
			router.Command("/add", s.handleContributorReposAdd)
			router.Command("/remove", s.handleContributorReposRemove)
			router.Command("/list", s.handleContributorReposList)
		})
	})
	router.Route("/docs", func(router handler.Router) {
		router.Command("/", s.handleDocs)
		router.Autocomplete("/", s.handleDocsAutocomplete)
	})
	router.Component("/docs_action", s.handleDocsAction)
	router.Component("/eval/rerun/{message_id}", s.handleEvalRerun)
	router.Component("/eval/delete", s.handleEvalDelete)
	router.Command("/eval", s.handleEval)
	router.Command("/info", s.handleInfo)
	router.Command("/ping", s.handlePing)
	router.Route("/tag", func(router handler.Router) {
		router.Command("/", s.handleTag)
		router.Autocomplete("/", s.handleTagListAutocomplete(false))
	})
	router.Route("/tags", func(router handler.Router) {
		router.Command("/create", s.handleCreateTag)
		router.Command("/edit", s.handleEditTag)
		router.Command("/delete", s.handleDeleteTag)
		router.Command("/info", s.handleTagInfo)
		router.Command("/list", s.handleListTags)
		router.Autocomplete("/edit", s.handleTagListAutocomplete(true))
		router.Autocomplete("/delete", s.handleTagListAutocomplete(true))
		router.Autocomplete("/info", s.handleTagListAutocomplete(false))
		router.Autocomplete("/list", s.handleTagListAutocomplete(false))
	})
	router.Command("/close-ticket", s.handleCloseTicket)
}

func (s *Server) SyncCommands(guildIDs []snowflake.ID) error {
	return handler.SyncCommands(s.Client, commands, guildIDs)
}

func (s *Server) registerLinkedRoles(ctx context.Context) error {
	repos, err := s.DB.ListContributorRepos(ctx)
	if err != nil {
		return err
	}
	linkedRoles := make([]discord.ApplicationRoleConnectionMetadata, 0, len(repos))
	for _, repo := range repos {
		linkedRoles = append(linkedRoles, discord.ApplicationRoleConnectionMetadata{
			Type:        discord.ApplicationRoleConnectionMetadataTypeIntegerGreaterThanOrEqual,
			Key:         strings.ReplaceAll(repo, "/", "_") + "_contributions",
			Name:        "Contributions to " + repo,
			Description: "The amount of contributions you have made to " + repo + " project",
		})
	}
	_, err = s.Client.Rest.UpdateApplicationRoleConnectionMetadata(s.Client.ApplicationID, linkedRoles)
	return err
}

func (s *Server) Start(ctx context.Context) error {
	if err := s.Client.OpenGateway(ctx); err != nil {
		return err
	}
	if err := s.Client.OpenHTTPServer(); err != nil {
		return err
	}
	contributorCtx, cancel := context.WithCancel(context.Background())
	s.contributorCancel = cancel
	go s.refreshContributorRoles(contributorCtx)
	return nil
}

func (s *Server) Stop() {
	if s.contributorCancel != nil {
		s.contributorCancel()
	}
	if s.Client != nil {
		s.Client.Close(context.Background())
	}
	if s.DB != nil {
		s.DB.Close()
	}
}

func (s *Server) onReady(_ *events.Ready) {
	slog.Info("butler ready")
	if err := s.Client.SetPresence(context.Background(),
		gateway.WithListeningActivity("you in DMs"),
		gateway.WithOnlineStatus(discord.OnlineStatusOnline),
	); err != nil {
		slog.Error("failed to set presence", slog.Any("err", err))
	}
}

func (s *Server) refreshContributorRoles(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			slog.Info("refreshing contributor roles")
			if err := s.updateContributorRoles(ctx); err != nil {
				slog.Error("failed to update contributor roles", slog.Any("err", err))
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *Server) updateContributorRoles(ctx context.Context) error {
	contributors, err := s.DB.ListContributors(ctx)
	if err != nil {
		return err
	}
	repos, err := s.DB.ListContributorRepos(ctx)
	if err != nil {
		return err
	}

	contributorRepos := map[string][]*github.Contributor{}
	for _, repo := range repos {
		values := strings.SplitN(repo, "/", 2)
		githubContributors, _, listErr := s.GitHub.Repositories.ListContributors(ctx, values[0], values[1], nil)
		if listErr != nil {
			return listErr
		}
		contributorRepos[repo] = githubContributors
	}

	for _, contributor := range contributors {
		metadata, metaErr := s.contributorMetadata(contributor.Username, contributorRepos)
		if metaErr != nil {
			slog.Error("failed to get contributor metadata", slog.String("username", contributor.Username), slog.Any("err", metaErr))
			continue
		}
		session, sessionErr := s.contributorSession(ctx, contributor)
		if sessionErr != nil {
			slog.Error("failed to refresh contributor session", slog.String("username", contributor.Username), slog.Any("err", sessionErr))
			continue
		}
		if _, err = s.OAuth2.UpdateApplicationRoleConnection(session, s.Client.ApplicationID, discord.ApplicationRoleConnectionUpdate{
			Metadata: &metadata,
		}); err != nil {
			slog.Error("failed to update contributor roles", slog.String("username", contributor.Username), slog.Any("err", err))
		}
	}
	return nil
}

func (s *Server) contributorSession(ctx context.Context, contributor sqlc.Contributor) (oauth2.Session, error) {
	session := oauth2.Session{
		AccessToken:  contributor.AccessToken,
		RefreshToken: contributor.RefreshToken,
		Scopes:       []discord.OAuth2Scope(contributor.Scopes),
		TokenType:    discord.TokenType(contributor.TokenType),
		Expiration:   contributor.Expiration,
	}
	refreshed, err := s.OAuth2.VerifySession(session, rest.WithCtx(ctx))
	if err != nil {
		return oauth2.Session{}, err
	}
	if refreshed.AccessToken == session.AccessToken {
		return session, nil
	}
	scopes := refreshed.Scopes
	if len(scopes) == 0 {
		scopes = session.Scopes
	}
	err = s.DB.CreateContributor(ctx, sqlc.CreateContributorParams{
		Username:     contributor.Username,
		AccessToken:  refreshed.AccessToken,
		RefreshToken: refreshed.RefreshToken,
		TokenType:    string(refreshed.TokenType),
		Scopes:       dbtype.OAuthScopes(scopes),
		Expiration:   refreshed.Expiration,
	})
	if err != nil {
		return oauth2.Session{}, err
	}
	return refreshed, nil
}
