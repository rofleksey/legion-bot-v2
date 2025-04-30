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

func InitAppTwitchClient(cfg *config.Config) (*helix.Client, error) {
	client, err := helix.NewClient(&helix.Options{
		ClientID:     cfg.Auth.ClientID,
		ClientSecret: cfg.Auth.ClientSecret,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create twitch client: %v", err)
	}

	resp, err := client.RequestAppAccessToken([]string{})
	if err != nil {
		return nil, fmt.Errorf("failed to request app access token: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to request app access token: invalid status code %d: %s (%s)", resp.StatusCode, resp.Error, resp.ErrorMessage)
	}

	client.SetAppAccessToken(resp.Data.AccessToken)

	return client, nil
}

func InitTwitchClients(clientID, accessToken string) (*twitch.Client, *helix.Client, error) {
	ircClient := twitch.NewClient(clientID, "oauth:"+accessToken)

	helixClient, err := helix.NewClient(&helix.Options{
		ClientID:        clientID,
		UserAccessToken: accessToken,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create helix client: %w", err)
	}

	return ircClient, helixClient, nil
}
