package butler

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo-butler/db"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/httpserver"
	"github.com/disgoorg/disgo/webhook"
	"github.com/disgoorg/log"
	"github.com/google/go-github/v44/github"
	"github.com/hhhapz/doc"
	"github.com/hhhapz/doc/godocs"
)

func New(version string, config Config) *Butler {
	return &Butler{
		Config:     config,
		Commands:   map[string]Command{},
		Components: map[string]Component{},
		Webhooks:   map[string]webhook.Client{},
		Version:    version,
	}
}

type Butler struct {
	Client       bot.Client
	Mux          *http.ServeMux
	GitHubClient *github.Client
	Commands     map[string]Command
	Components   map[string]Component
	DocClient    *doc.CachedSearcher
	DB           db.DB
	Config       Config
	Webhooks     map[string]webhook.Client
	Version      string
}

func (b *Butler) SetupRoutes(routes http.Handler) {
	b.Mux = http.NewServeMux()
	b.Mux.Handle("/", routes)
}

func (b *Butler) SetupBot() {
	var err error
	if b.Client, err = disgo.New(b.Config.Token,
		bot.WithGatewayConfigOpts(
			gateway.WithGatewayIntents(discord.GatewayIntentsAll),
			gateway.WithCompress(true),
			gateway.WithPresence(discord.GatewayMessageDataPresenceUpdate{
				Activities: []discord.Activity{
					{
						Name: "loading...",
						Type: discord.ActivityTypeGame,
					},
				},
				Status: discord.OnlineStatusDND,
			}),
		),
		bot.WithCacheConfigOpts(cache.WithCacheFlags(cache.FlagGuilds)),
		bot.WithEventListenerFunc(b.OnReady),
		bot.WithEventListenerFunc(b.OnApplicationCommandInteraction),
		bot.WithEventListenerFunc(b.OnComponentInteraction),
		bot.WithEventListenerFunc(b.OnAutocompleteInteraction),
		bot.WithHTTPServerConfigOpts(
			httpserver.WithServeMux(b.Mux),
			httpserver.WithAddress(b.Config.InteractionsConfig.Address),
			httpserver.WithURL(b.Config.InteractionsConfig.URL),
			httpserver.WithPublicKey(b.Config.InteractionsConfig.PublicKey),
		),
	); err != nil {
		log.Errorf("Failed to start bot: %s", err)
	}
	b.GitHubClient = github.NewClient(b.Client.Rest().HTTPClient())
	b.DocClient = doc.WithCache(doc.New(b.Client.Rest().HTTPClient(), godocs.Parser))
	for _, module := range b.Config.Docs.Aliases {
		_, _ = b.DocClient.Search(context.TODO(), module)
	}
}

func (b *Butler) SetupDB(shouldSyncDBTables bool) {
	var err error
	if b.DB, err = db.SetupDatabase(shouldSyncDBTables, b.Config.Database); err != nil {
		log.Fatalf("Failed to setup database: %s", err)
	}
}

func (b *Butler) StartAndBlock() {
	if err := b.Client.ConnectGateway(context.TODO()); err != nil {
		log.Errorf("Failed to connect to gateway: %s", err)
	}

	if err := b.Client.StartHTTPServer(); err != nil {
		log.Errorf("Failed to start http server: %s", err)
	}

	log.Info("Client is running. Press CTRL-C to exit.")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-s
	log.Info("Shutting down...")
}

func (b *Butler) OnReady(_ *events.ReadyEvent) {
	log.Infof("Butler ready")
	if err := b.Client.SetPresence(context.TODO(), discord.GatewayMessageDataPresenceUpdate{
		Activities: []discord.Activity{
			{
				Name: "you",
				Type: discord.ActivityTypeListening,
			},
		},
		Status: discord.OnlineStatusOnline,
	}); err != nil {
		log.Errorf("Failed to set presence: %s", err)
	}
}
