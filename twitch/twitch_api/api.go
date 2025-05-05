package twitch_api

import (
	"context"
	"fmt"
	"github.com/gempir/go-twitch-irc/v4"
	"github.com/nicklaw5/helix/v2"
	"github.com/samber/do"
	"legion-bot-v2/config"
	"log/slog"
	"net/http"
	"time"
)

var appScopes = []string{
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
}

type TwitchApi struct {
	cfg        *config.Config
	userClient *helix.Client
	appClient  *helix.Client
	ircClient  *twitch.Client
}

func NewTwitchApi(di *do.Injector) (*TwitchApi, error) {
	cfg := do.MustInvoke[*config.Config](di)

	userClient, err := newUserClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("error getting user client: %v", err)
	}

	appClient, err := newAppClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("error getting app client: %v", err)
	}

	ircClient, err := newIrcClient(cfg, userClient.GetUserAccessToken())
	if err != nil {
		return nil, fmt.Errorf("error getting IRC client: %v", err)
	}

	return &TwitchApi{
		cfg:        cfg,
		userClient: userClient,
		appClient:  appClient,
		ircClient:  ircClient,
	}, nil
}

func (a *TwitchApi) UserClient() *helix.Client {
	return a.userClient
}

func (a *TwitchApi) AppClient() *helix.Client {
	return a.appClient
}

func (a *TwitchApi) IrcClient() *twitch.Client {
	return a.ircClient
}

func (a *TwitchApi) GetUserToken() string {
	return a.userClient.GetUserAccessToken()
}

func (a *TwitchApi) GetAppToken() string {
	return a.appClient.GetAppAccessToken()
}

func (a *TwitchApi) Run(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.refreshTokens()
		}
	}
}

func (a *TwitchApi) refreshTokens() {
	slog.Debug("Twitch token refresh job started")
	defer slog.Debug("Twitch token refresh job finished")

	if err := a.refreshUserToken(); err != nil {
		slog.Error("Failed to refresh twitch user token",
			slog.Any("error", err),
		)
	}

	if err := a.refreshAppToken(); err != nil {
		slog.Error("Failed to refresh twitch app token",
			slog.Any("error", err),
		)
	}

	a.ircClient.SetIRCToken("oauth:" + a.userClient.GetUserAccessToken())
}

func (a *TwitchApi) refreshUserToken() error {
	resp, err := a.userClient.RefreshUserAccessToken(a.cfg.Twitch.RefreshToken)
	if err != nil {
		return fmt.Errorf("failed to refresh user access token: %v", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to refresh app access token: invalid status code %d: %s (%s)", resp.StatusCode, resp.Error, resp.ErrorMessage)
	}

	a.userClient.SetUserAccessToken(resp.Data.AccessToken)

	return nil
}

func (a *TwitchApi) refreshAppToken() error {
	resp, err := a.appClient.RequestAppAccessToken(appScopes)
	if err != nil {
		return fmt.Errorf("failed to refresh app access token: %v", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed to refresh app access token: invalid status code %d: %s (%s)", resp.StatusCode, resp.Error, resp.ErrorMessage)
	}

	a.appClient.SetAppAccessToken(resp.Data.AccessToken)

	return nil
}

func newUserClient(cfg *config.Config) (*helix.Client, error) {
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
		return nil, fmt.Errorf("failed to create twitch client: %v", err)
	}

	resp, err := client.RefreshUserAccessToken(cfg.Twitch.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh user access token: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to request app access token: invalid status code %d: %s (%s)", resp.StatusCode, resp.Error, resp.ErrorMessage)
	}

	client.SetUserAccessToken(resp.Data.AccessToken)

	return client, nil
}

func newAppClient(cfg *config.Config) (*helix.Client, error) {
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

	resp, err := client.RequestAppAccessToken(appScopes)
	if err != nil {
		return nil, fmt.Errorf("failed to request app access token: %v", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("failed to request app access token: invalid status code %d: %s (%s)", resp.StatusCode, resp.Error, resp.ErrorMessage)
	}

	client.SetAppAccessToken(resp.Data.AccessToken)

	return client, nil
}

func newIrcClient(cfg *config.Config, userAccessToken string) (*twitch.Client, error) {
	ircClient := twitch.NewClient(cfg.Twitch.ClientID, "oauth:"+userAccessToken)

	return ircClient, nil
}
