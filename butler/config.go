package butler

import (
	"encoding/json"
	"os"
)

func LoadConfig() (*Config, error) {
	file, err := os.OpenFile("config.json", os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err = json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

type Config struct {
	Token string `json:"token"`

	GithubWebhookSecret string `json:"github_webhook_secret"`

	InteractionsConfig InteractionsConfig `json:"interactions"`
}

type InteractionsConfig struct {
	URL       string `json:"url"`
	Port      string `json:"port"`
	PublicKey string `json:"public_key"`
}
