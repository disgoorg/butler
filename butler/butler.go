package butler

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/DisgoOrg/disgo/core"
	"github.com/DisgoOrg/disgo/core/bot"
	"github.com/DisgoOrg/disgo/core/events"
	"github.com/DisgoOrg/disgo/discord"
	"github.com/DisgoOrg/disgo/gateway"
	"github.com/DisgoOrg/disgo/httpserver"
	"github.com/DisgoOrg/log"
	"github.com/DisgoOrg/snowflake"
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
	Bot          *core.Bot
	Mux          *http.ServeMux
	GitHubClient *github.Client
	Commands     map[snowflake.Snowflake]Command
	Components   *Components
	DocClient    *doc.CachedSearcher
	DB           *bun.DB
	Config       Config
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
		cmds []core.ApplicationCommand
		err  error
	)
	if b.Config.DevMode {
		cmds, err = b.Bot.SetGuildCommands(b.Config.DevGuildID, commandCreates)
	} else {
		cmds, err = b.Bot.SetCommands(commandCreates)
	}
	if err != nil {
		b.Bot.Logger.Error("Failed to set commands: ", err)
	}
	for i, cmd := range cmds {
		b.Commands[cmd.ID()] = commands[i]
	}
}

func (b *Butler) SetupComponents(components map[string]func(b *Butler, data []string, e *events.ComponentInteractionEvent)) {
	for action, handler := range components {
		b.Components.Add(action, handler, 0)
	}
}

func (b *Butler) SetupBot() {
	var err error
	if b.Bot, err = bot.New(b.Config.Token,
		bot.WithGatewayOpts(
			gateway.WithGatewayIntents(discord.GatewayIntentsAll),
			gateway.WithCompress(true),
			gateway.WithPresence(discord.UpdatePresenceCommandData{
				Activities: []discord.Activity{
					{
						Name: "loading...",
						Type: discord.ActivityTypeGame,
					},
				},
				Status: discord.OnlineStatusDND,
			}),
		),
		bot.WithCacheOpts(core.WithCacheFlags(core.CacheFlagGuilds)),
		bot.WithEventListeners(b),
		bot.WithHTTPServerOpts(
			httpserver.WithServeMux(b.Mux),
			httpserver.WithPort(b.Config.InteractionsConfig.Port),
			httpserver.WithURL(b.Config.InteractionsConfig.URL),
			httpserver.WithPublicKey(b.Config.InteractionsConfig.PublicKey),
		),
	); err != nil {
		log.Errorf("Failed to start bot: %s", err)
	}
	b.GitHubClient = github.NewClient(b.Bot.RestServices.HTTPClient())
	b.DocClient = doc.WithCache(doc.New(b.Bot.RestServices.HTTPClient(), godocs.Parser))
}

func (b *Butler) StartAndBlock() {
	if err := b.Bot.ConnectGateway(context.TODO()); err != nil {
		log.Errorf("Failed to connect to gateway: %s", err)
	}

	if err := b.Bot.StartHTTPServer(); err != nil {
		log.Errorf("Failed to start http server: %s", err)
	}

	log.Info("Bot is running. Press CTRL-C to exit.")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-s
	log.Info("Shutting down...")
}

func (b *Butler) OnEvent(event core.Event) {
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
	if err := b.Bot.SetPresence(context.TODO(), discord.UpdatePresenceCommandData{
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
