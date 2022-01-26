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
)

func New(config Config) *Bot {
	b := &Bot{
		Config: config,
	}
	b.ListenerAdapter = &events.ListenerAdapter{
		OnReady: b.OnReady,
	}
	return b
}

type Bot struct {
	*events.ListenerAdapter
	Bot    *core.Bot
	Mux    *http.ServeMux
	Config Config
}

func (b *Bot) SetupHTTPServer() {
	b.Mux = http.NewServeMux()
	b.Mux.HandleFunc("/github", b.HandleGithubWebhook)
	b.Mux.HandleFunc("/login", b.HandleDiscordLogin)
}

func (b *Bot) StartAndBlock() {
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
		bot.WithCacheOpts(
			core.WithCacheFlags(
				core.CacheFlagGuilds,
			),
		),
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

	if err = b.Bot.ConnectGateway(context.TODO()); err != nil {
		log.Errorf("Failed to connect to gateway: %s", err)
	}

	if err = b.Bot.StartHTTPServer(); err != nil {
		log.Errorf("Failed to start http server: %s", err)
	}

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-s
	log.Info("Shutting down...")
}

func (b *Bot) OnReady(_ *events.ReadyEvent) {
	log.Infof("Bot ready")
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
