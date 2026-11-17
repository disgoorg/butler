package server

import (
	"log/slog"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/disgoorg/snowflake/v2"

	"github.com/disgoorg/disgo-butler/server/database"
)

func LoadConfig(path string) (Config, []toml.Key, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, nil, err
	}
	defer file.Close()

	var cfg Config
	meta, err := toml.NewDecoder(file).Decode(&cfg)
	if err != nil {
		return Config{}, nil, err
	}
	return cfg, meta.Undecoded(), nil
}

type Config struct {
	DevMode             bool               `toml:"dev_mode"`
	GuildID             snowflake.ID       `toml:"guild_id"`
	Token               string             `toml:"token"`
	Secret              string             `toml:"secret"`
	BaseURL             string             `toml:"base_url"`
	GithubWebhookSecret string             `toml:"github_webhook_secret"`
	Log                 LogConfig          `toml:"log"`
	Database            database.Config    `toml:"database"`
	Interactions        InteractionsConfig `toml:"interactions"`
	ModMail             ModMailConfig      `toml:"mod_mail"`
}

type LogConfig struct {
	Level     slog.Level `toml:"level"`
	Format    string     `toml:"format"`
	AddSource bool       `toml:"add_source"`
}

type InteractionsConfig struct {
	URL       string `toml:"url"`
	Address   string `toml:"address"`
	PublicKey string `toml:"public_key"`
}

type ModMailConfig struct {
	RoleID       snowflake.ID `toml:"role_id"`
	ChannelID    snowflake.ID `toml:"channel_id"`
	WebhookID    snowflake.ID `toml:"webhook_id"`
	WebhookToken string       `toml:"webhook_token"`
}
