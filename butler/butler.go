package butler

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/httpserver"
	"github.com/disgoorg/disgo/webhook"
	"github.com/disgoorg/log"
	"github.com/disgoorg/snowflake"
	"github.com/google/go-github/github"
	"github.com/hhhapz/doc"
	"github.com/hhhapz/doc/godocs"
	"github.com/uptrace/bun"
)

func New(config Config) *Butler {
	return &Butler{
		Config:     config,
		Commands:   map[snowflake.Snowflake]Command{},
		Components: NewComponents(),
	}
}

type Butler struct {
	Client       bot.Client
	Mux          *http.ServeMux
	GitHubClient *github.Client
	Commands     map[snowflake.Snowflake]Command
	Components   *Components
	DocClient    *doc.CachedSearcher
	DB           *bun.DB
	Config       Config
	Webhooks     map[string]webhook.Client
}

func (b *Butler) SetupHTTPHandlers(handlers map[string]HTTPHandleFunc) {
	b.Mux = http.NewServeMux()
	for pattern, handler := range handlers {
		b.Mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			handler(b, w, r)
		})
	}
}

func (b *Butler) SetupCommands(commands []Command) {
	commandCreates := make([]discord.ApplicationCommandCreate, len(commands))
	for i := range commands {
		commandCreates[i] = commands[i].Create
	}
	var (
		cmds []discord.ApplicationCommand
		err  error
	)
	if b.Config.DevMode {
		cmds, err = b.Client.Rest().Applications().SetGuildCommands(b.Client.ApplicationID(), b.Config.DevGuildID, commandCreates)
	} else {
		cmds, err = b.Client.Rest().Applications().SetGlobalCommands(b.Client.ApplicationID(), commandCreates)
	}
	if err != nil {
		b.Client.Logger().Error("Failed to set commands: ", err)
	}
	for i, cmd := range cmds {
		b.Commands[cmd.ID()] = commands[i]
	}
}

func (b *Butler) SetupComponents(components map[string]func(b *Butler, data []string, e *events.ComponentInteractionEvent) error) {
	for action, handler := range components {
		b.Components.Add(action, handler, 0)
	}
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
		bot.WithEventListeners(b),
		bot.WithHTTPServerConfigOpts(
			httpserver.WithServeMux(b.Mux),
			httpserver.WithAddress(b.Config.InteractionsConfig.Address),
			httpserver.WithURL(b.Config.InteractionsConfig.URL),
			httpserver.WithPublicKey(b.Config.InteractionsConfig.PublicKey),
		),
	); err != nil {
		log.Errorf("Failed to start bot: %s", err)
	}
	b.GitHubClient = github.NewClient(b.Client.Rest().RestClient().HTTPClient())
	b.DocClient = doc.WithCache(doc.New(b.Client.Rest().RestClient().HTTPClient(), godocs.Parser))
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

func (b *Butler) OnEvent(event bot.Event) {
	switch e := event.(type) {
	case *events.ReadyEvent:
		b.OnReady()

	case *events.ApplicationCommandInteractionEvent:
		b.OnApplicationCommandInteraction(e)

	case *events.ComponentInteractionEvent:
		b.OnComponentInteraction(e)

	case *events.AutocompleteInteractionEvent:
		b.OnAutocompleteInteraction(e)
	}
}

func (b *Butler) OnReady() {
	log.Infof("Butler ready")
	if err := b.Client.SetPresence(context.TODO(), discord.GatewayMessageDataPresenceUpdate{
		Activities: []discord.Activity{
			{
				Name: "to you",
				Type: discord.ActivityTypeListening,
			},
		},
		Status: discord.OnlineStatusOnline,
	}); err != nil {
		log.Errorf("Failed to set presence: %s", err)
	}
}
