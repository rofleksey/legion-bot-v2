package util

import (
	"fmt"
	"github.com/gempir/go-twitch-irc/v4"
	"github.com/nicklaw5/helix/v2"
	"legion-bot-v2/config"
	"net/http"
	"time"
)

func FetchTwitchUserAccessToken(cfg *config.Config) (string, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:    10,
			IdleConnTimeout: 30 * time.Second,
		},
		Timeout: 10 * time.Second,
	}

	client, err := helix.NewClient(&helix.Options{
		ClientID:     cfg.Twitch.ClientID,
		ClientSecret: cfg.Twitch.ClientSecret,
		RefreshToken: cfg.Twitch.RefreshToken,
		HTTPClient:   httpClient,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create twitch client: %v", err)
	}

	resp, err := client.RefreshUserAccessToken(cfg.Twitch.RefreshToken)
	if err != nil {
		return "", fmt.Errorf("failed to refresh user access token: %v", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to request app access token: invalid status code %d: %s (%s)", resp.StatusCode, resp.Error, resp.ErrorMessage)
	}

	return resp.Data.AccessToken, nil
}

func InitAppTwitchClient(cfg *config.Config, userAccessToken string) (*helix.Client, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:    10,
			IdleConnTimeout: 30 * time.Second,
		},
		Timeout: 10 * time.Second,
	}

	client, err := helix.NewClient(&helix.Options{
		ClientID:     cfg.Twitch.ClientID,
		ClientSecret: cfg.Twitch.ClientSecret,
		HTTPClient:   httpClient,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create twitch client: %v", err)
	}

	resp, err := client.RequestAppAccessToken([]string{
		"channel:read:guest_star",
		"channel:manage:guest_star",
		"moderator:read:guest_star",
		"moderator:manage:guest_star",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to request app access token: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to request app access token: invalid status code %d: %s (%s)", resp.StatusCode, resp.Error, resp.ErrorMessage)
	}

	client.SetAppAccessToken(resp.Data.AccessToken)
	client.SetUserAccessToken(userAccessToken)

	return client, nil
}

func InitTwitchClients(cfg *config.Config, userAccessToken string) (*twitch.Client, *helix.Client, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:    10,
			IdleConnTimeout: 30 * time.Second,
		},
		Timeout: 10 * time.Second,
	}

	helixClient, err := helix.NewClient(&helix.Options{
		ClientID:        cfg.Twitch.ClientID,
		UserAccessToken: userAccessToken,
		HTTPClient:      httpClient,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create helix client: %w", err)
	}

	ircClient := twitch.NewClient(cfg.Twitch.ClientID, "oauth:"+userAccessToken)

	return ircClient, helixClient, nil
}
