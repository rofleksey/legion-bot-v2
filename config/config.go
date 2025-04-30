package config

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type Config struct {
	BaseURL string `yaml:"baseUrl" validate:"required"`

	Chat struct {
		ClientID      string `yaml:"clientId" validate:"required"`
		RefreshToken  string `yaml:"refreshToken" validate:"required"`
		WebHookSecret string `yaml:"webhookSecret" validate:"required"`
	} `yaml:"chat" validate:"required"`

	Auth struct {
		ClientID     string `yaml:"clientId" validate:"required"`
		ClientSecret string `yaml:"clientSecret" validate:"required"`
		JwtSecret    string `yaml:"jwtSecret" validate:"required"`
		RedirectURL  string `yaml:"redirectUrl" validate:"required"`
	} `yaml:"auth" validate:"required"`

	Telegram struct {
		Token  string `yaml:"token" validate:"required"`
		ChatID string `yaml:"chatId" validate:"required"`
	}

	Yandex struct {
		ServiceAccountID string `yaml:"serviceAccountId" validate:"required"`
		FolderID         string `yaml:"folderId" validate:"required"`
		KeyID            string `yaml:"keyId" validate:"required"`
		Key              string `yaml:"key" validate:"required"`
	}
}

func Load() (*Config, error) {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}
