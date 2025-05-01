package util

import (
	"encoding/json"
	"fmt"
	"github.com/gempir/go-twitch-irc/v4"
	"github.com/nicklaw5/helix/v2"
	"legion-bot-v2/config"
	"net/http"
	"time"
)

func FetchTwitchAccessToken(refreshToken string) (string, error) {
	client := &http.Client{
		Timeout: time.Second * 30,
	}

	req, err := http.NewRequest("GET", RefreshURL+"/"+refreshToken, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to refresh token: %s", resp.Status)
	}

	var result struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Token, nil
}

func InitAppTwitchClient(cfg *config.Config, accessToken string) (*helix.Client, error) {
	client, err := helix.NewClient(&helix.Options{
		ClientID:     cfg.Auth.ClientID,
		ClientSecret: cfg.Auth.ClientSecret,
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
	client.SetUserAccessToken(accessToken)

	return client, nil
}

func InitTwitchClients(cfg *config.Config, accessToken string) (*twitch.Client, *helix.Client, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:    10,
			IdleConnTimeout: 30 * time.Second,
		},
		Timeout: 10 * time.Second,
	}

	helixClient, err := helix.NewClient(&helix.Options{
		ClientID:        cfg.Chat.ClientID,
		UserAccessToken: accessToken,
		HTTPClient:      httpClient,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create helix client: %w", err)
	}

	ircClient := twitch.NewClient(cfg.Auth.ClientID, "oauth:"+accessToken)

	return ircClient, helixClient, nil
}
