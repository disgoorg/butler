package butler

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/DisgoOrg/disgo/webhook"
	"github.com/DisgoOrg/log"
	"github.com/DisgoOrg/snowflake"
)

func LoadConfig() (*Config, error) {
	file, err := os.Open("config.json")
	if os.IsNotExist(err) {
		if file, err = os.Create("config.json"); err != nil {
			return nil, err
		}
		var data []byte
		if data, err = json.Marshal(Config{}); err != nil {
			return nil, err
		}
		if _, err = file.Write(data); err != nil {
			return nil, err
		}
		return nil, errors.New("config.json not found, created new one")
	} else if err != nil {
		return nil, err
	}

	var cfg Config
	if err = json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

type Config struct {
	DevMode    bool                `json:"dev_mode"`
	DevGuildID snowflake.Snowflake `json:"dev_guild_id"`
	LogLevel   log.Level           `json:"log_level"`
	Token      string              `json:"token"`

	DocsConfig DocsConfig `json:"docs_config"`

	GithubWebhookSecret  string                         `json:"github_webhook_secret"`
	GithubReleasesConfig map[string]GithubReleaseConfig `json:"releasers"`

	InteractionsConfig InteractionsConfig `json:"interactions"`
}

type DocsConfig struct {
	Aliases map[string]string `json:"aliases"`
}

type InteractionsConfig struct {
	URL       string `json:"url"`
	Port      string `json:"port"`
	PublicKey string `json:"public_key"`
}

type GithubReleaseConfig struct {
	WebhookID     snowflake.Snowflake `json:"webhook_id"`
	WebhookToken  string              `json:"webhook_token"`
	PingRole      snowflake.Snowflake `json:"ping_role"`
	WebhookClient *webhook.Client     `json:"-"`
}
