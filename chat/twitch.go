package chat

import (
	"context"
	"github.com/gempir/go-twitch-irc/v4"
	"github.com/nicklaw5/helix/v2"
	"golang.org/x/sync/semaphore"
	"legion-bot-v2/util"
	"log/slog"
	"sync"
	"time"
)

var _ Actions = (*TwitchActions)(nil)

type TwitchActions struct {
	ircClient   *twitch.Client
	helixClient *helix.Client
	queues      map[string]*semaphore.Weighted
	mu          sync.Mutex
}

func NewTwitchActions(ircClient *twitch.Client, helixClient *helix.Client) *TwitchActions {
	return &TwitchActions{
		ircClient:   ircClient,
		helixClient: helixClient,
		queues:      make(map[string]*semaphore.Weighted),
	}
}

func (t *TwitchActions) getQueue(channel string) *semaphore.Weighted {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.queues[channel]; !exists {
		t.queues[channel] = semaphore.NewWeighted(1)
	}
	return t.queues[channel]
}

func (t *TwitchActions) SendMessage(channel, text string) {
	queue := t.getQueue(channel)

	go func() {
		if err := queue.Acquire(context.Background(), 1); err != nil {
			slog.Error("Failed to acquire semaphore",
				slog.String("channel", channel),
				slog.Any("error", err),
			)
			return
		}
		defer queue.Release(1)

		slog.Info("<<<",
			slog.String("channel", channel),
			slog.String("text", text),
		)

		t.ircClient.Say(channel, text)
	}()
}

func (t *TwitchActions) TimeoutUser(channel, username string, duration time.Duration, reason string) {
	queue := t.getQueue(channel)

	go func() {
		if err := queue.Acquire(context.Background(), 1); err != nil {
			slog.Error("Failed to acquire semaphore",
				slog.String("channel", channel),
				slog.Any("error", err),
			)
			return
		}
		defer queue.Release(1)

		slog.Info("User has been timeout",
			slog.String("channel", channel),
			slog.String("username", username),
			slog.Duration("duration", duration),
			slog.String("reason", reason),
		)

		botResp, err := t.helixClient.GetUsers(&helix.UsersParams{
			Logins: []string{util.BotUsername},
		})
		if err != nil || len(botResp.Data.Users) == 0 {
			slog.Error("Error getting bot user",
				slog.String("channel", channel),
				slog.Any("error", err),
			)
			return
		}
		botUser := botResp.Data.Users[0]

		channelResp, err := t.helixClient.GetUsers(&helix.UsersParams{
			Logins: []string{channel},
		})
		if err != nil || len(channelResp.Data.Users) == 0 {
			slog.Error("Error getting channel user",
				slog.String("channel", channel),
				slog.Any("error", err),
			)
			return
		}
		channelUser := channelResp.Data.Users[0]

		// Get user to ban ID
		userResp, err := t.helixClient.GetUsers(&helix.UsersParams{
			Logins: []string{username},
		})
		if err != nil || len(userResp.Data.Users) == 0 {
			slog.Error("Error getting user to ban",
				slog.String("channel", channel),
				slog.Any("error", err),
			)
			return
		}
		banUser := userResp.Data.Users[0]

		// Set the user access token for the bot (assuming you've previously authenticated)
		// Note: You'll need to handle OAuth properly in your application
		t.helixClient.SetUserAccessToken(botUser.ID)

		// Perform the ban
		banResp, err := t.helixClient.BanUser(&helix.BanUserParams{
			BroadcasterID: channelUser.ID,
			ModeratorId:   botUser.ID,
			Body: helix.BanUserRequestBody{
				Duration: int(duration.Seconds()),
				Reason:   reason,
				UserId:   banUser.ID,
			},
		})
		if err != nil {
			slog.Error("Error banning user",
				slog.String("channel", channel),
				slog.Any("error", err),
			)
			return
		}

		if banResp.StatusCode >= 400 {
			slog.Error("Ban API error",
				slog.String("channel", channel),
				slog.Any("error", err),
			)
		}
	}()
}

func (t *TwitchActions) UnbanUser(channel, username string) {
	queue := t.getQueue(channel)

	go func() {
		if err := queue.Acquire(context.Background(), 1); err != nil {
			slog.Error("Failed to acquire semaphore",
				slog.String("channel", channel),
				slog.Any("error", err),
			)
			return
		}
		defer queue.Release(1)

		slog.Info("User has been unbanned",
			slog.String("channel", channel),
			slog.String("username", username),
		)

		// Get bot user ID
		botResp, err := t.helixClient.GetUsers(&helix.UsersParams{
			Logins: []string{util.BotUsername},
		})
		if err != nil || len(botResp.Data.Users) == 0 {
			slog.Error("Error getting bot user",
				slog.String("channel", channel),
				slog.Any("error", err),
			)
			return
		}
		botUser := botResp.Data.Users[0]

		// Get channel user ID
		channelResp, err := t.helixClient.GetUsers(&helix.UsersParams{
			Logins: []string{channel},
		})
		if err != nil || len(channelResp.Data.Users) == 0 {
			slog.Error("Error getting channel user",
				slog.String("channel", channel),
				slog.Any("error", err),
			)
			return
		}
		channelUser := channelResp.Data.Users[0]

		// Get user to unban ID
		userResp, err := t.helixClient.GetUsers(&helix.UsersParams{
			Logins: []string{username},
		})
		if err != nil || len(userResp.Data.Users) == 0 {
			slog.Error("Error getting user to unban",
				slog.String("channel", channel),
				slog.Any("error", err),
			)
			return
		}
		banUser := userResp.Data.Users[0]

		// Set the user access token for the bot
		t.helixClient.SetUserAccessToken(botUser.ID)

		// Perform the unban
		unbanResp, err := t.helixClient.UnbanUser(&helix.UnbanUserParams{
			BroadcasterID: channelUser.ID,
			ModeratorID:   botUser.ID,
			UserID:        banUser.ID,
		})
		if err != nil {
			slog.Error("Error unbanning user",
				slog.String("channel", channel),
				slog.Any("error", err),
			)
			return
		}

		if unbanResp.StatusCode >= 400 {
			slog.Error("Unban API error",
				slog.String("channel", channel),
				slog.Any("error", err),
			)
		}
	}()
}
