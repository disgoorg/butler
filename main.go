package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/disgoorg/snowflake/v2"

	"github.com/disgoorg/disgo-butler/server"
)

func main() {
	configPath := flag.String("config", "config.toml", "path to the config file")
	syncCommands := flag.Bool("sync-commands", false, "sync application commands on startup")
	flag.Parse()

	cfg, unknown, err := server.LoadConfig(*configPath)
	if err != nil {
		slog.Error("failed to load config", slog.String("path", *configPath), slog.Any("err", err))
		os.Exit(1)
	}
	setupLogger(cfg.Log)
	if len(unknown) > 0 {
		slog.Warn("unknown config keys", slog.Any("keys", unknown))
	}

	version := buildVersion()
	slog.Info("starting DisGo-Butler", slog.String("version", version))

	srv, err := server.New(context.Background(), cfg, version)
	if err != nil {
		slog.Error("failed to start", slog.Any("err", err))
		os.Exit(1)
	}
	defer srv.Stop()

	if *syncCommands {
		var guildIDs []snowflake.ID
		if cfg.DevMode {
			guildIDs = append(guildIDs, cfg.GuildID)
		}
		if err = srv.SyncCommands(guildIDs); err != nil {
			slog.Error("failed to sync commands", slog.Any("err", err))
		}
	}

	if err = srv.Start(context.Background()); err != nil {
		slog.Error("failed to open connections", slog.Any("err", err))
		os.Exit(1)
	}

	slog.Info("client is running")
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-signals
}

func buildVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}

	var revision, modified string
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			revision = setting.Value
		case "vcs.modified":
			modified = setting.Value
		}
	}
	if revision == "" {
		return "unknown"
	}
	if modified == "true" {
		return revision + "-dirty"
	}
	return revision
}

func setupLogger(cfg server.LogConfig) {
	options := &slog.HandlerOptions{
		AddSource: cfg.AddSource,
		Level:     cfg.Level,
	}
	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, options)
	} else {
		handler = slog.NewTextHandler(os.Stdout, options)
	}
	slog.SetDefault(slog.New(handler))
}
