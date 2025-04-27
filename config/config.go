package config

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Chat struct {
		ClientID     string `yaml:"clientId" validate:"required"`
		RefreshToken string `yaml:"refreshToken" validate:"required"`
	} `yaml:"chat" validate:"required"`

	Auth struct {
		ClientID     string `yaml:"clientId" validate:"required"`
		ClientSecret string `yaml:"clientSecret" validate:"required"`
		JwtSecret    string `yaml:"jwtSecret" validate:"required"`
		RedirectURL  string `yaml:"redirectUrl" validate:"required"`
	} `yaml:"auth" validate:"required"`
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
