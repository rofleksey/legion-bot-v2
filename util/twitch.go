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
		"analytics:read:extensions",
		"user:edit",
		"user:read:email",
		"clips:edit",
		"bits:read",
		"analytics:read:games",
		"user:edit:broadcast",
		"user:read:broadcast",
		"chat:read",
		"chat:edit",
		"channel:moderate",
		"channel:read:subscriptions",
		"whispers:read",
		"whispers:edit",
		"moderation:read",
		"channel:read:redemptions",
		"channel:edit:commercial",
		"channel:read:hype_train",
		"channel:read:stream_key",
		"channel:manage:extensions",
		"channel:manage:broadcast",
		"user:edit:follows",
		"channel:manage:redemptions",
		"channel:read:editors",
		"channel:manage:videos",
		"user:read:blocked_users",
		"user:manage:blocked_users",
		"user:read:subscriptions",
		"user:read:follows",
		"channel:manage:polls",
		"channel:manage:predictions",
		"channel:read:polls",
		"channel:read:predictions",
		"moderator:manage:automod",
		"channel:manage:schedule",
		"channel:read:goals",
		"moderator:read:automod_settings",
		"moderator:manage:automod_settings",
		"moderator:manage:banned_users",
		"moderator:read:blocked_terms",
		"moderator:manage:blocked_terms",
		"moderator:read:chat_settings",
		"moderator:manage:chat_settings",
		"channel:manage:raids",
		"moderator:manage:announcements",
		"moderator:manage:chat_messages",
		"user:manage:chat_color",
		"channel:manage:moderators",
		"channel:read:vips",
		"channel:manage:vips",
		"user:manage:whispers",
		"channel:read:charity",
		"moderator:read:chatters",
		"moderator:read:shield_mode",
		"moderator:manage:shield_mode",
		"moderator:read:shoutouts",
		"moderator:manage:shoutouts",
		"moderator:read:followers",
		"channel:read:guest_star",
		"channel:manage:guest_star",
		"moderator:read:guest_star",
		"moderator:manage:guest_star",
		"channel:bot",
		"user:bot",
		"user:read:chat",
		"channel:manage:ads",
		"channel:read:ads",
		"user:read:moderated_channels",
		"user:write:chat",
		"user:read:emotes",
		"moderator:read:unban_requests",
		"moderator:manage:unban_requests",
		"moderator:read:suspicious_users",
		"moderator:manage:warnings",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to request app access token: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to request app access token: invalid status code %d: %s (%s)", resp.StatusCode, resp.Error, resp.ErrorMessage)
	}

	client.SetAppAccessToken(resp.Data.AccessToken)

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
