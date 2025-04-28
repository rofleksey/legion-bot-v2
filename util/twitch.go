package util

import (
	"encoding/json"
	"fmt"
	"github.com/gempir/go-twitch-irc/v4"
	"github.com/nicklaw5/helix/v2"
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

func InitTwitchClients(clientID, accessToken string) (*twitch.Client, *helix.Client, error) {
	ircClient := twitch.NewClient(clientID, "oauth:"+accessToken)

	helixClient, err := helix.NewClient(&helix.Options{
		ClientID: clientID,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create helix client: %w", err)
	}

	helixClient.SetUserAccessToken(accessToken)

	return ircClient, helixClient, nil
}
