package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gempir/go-twitch-irc/v4"
	"github.com/nicklaw5/helix/v2"
	"io"
	"legion-bot-v2/config"
	"legion-bot-v2/taskq"
	"legion-bot-v2/util"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

var _ Actions = (*TwitchActions)(nil)

type TwitchActions struct {
	cfg         *config.Config
	accessToken string
	ircClient   *twitch.Client
	helixClient *helix.Client
	queues      map[string]*taskq.Queue
	mu          sync.Mutex
}

func NewTwitchActions(
	cfg *config.Config,
	accessToken string,
	ircClient *twitch.Client,
	helixClient *helix.Client,
) *TwitchActions {
	return &TwitchActions{
		cfg:         cfg,
		accessToken: accessToken,
		ircClient:   ircClient,
		helixClient: helixClient,
		queues:      make(map[string]*taskq.Queue),
	}
}

func (t *TwitchActions) getQueue(channel string) *taskq.Queue {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.queues[channel]; !exists {
		t.queues[channel] = taskq.New(1, 1, 1)
	}
	return t.queues[channel]
}

func (t *TwitchActions) Shutdown() {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, q := range t.queues {
		q.Shutdown()
	}
}

func (t *TwitchActions) getGuestStarSessionIsActive(broadcasterID, moderatorID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := "https://api.twitch.tv/helix/guest_star/session?broadcaster_id=" + broadcasterID + "&moderator_id=" + moderatorID

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create guest star session request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+t.accessToken)
	req.Header.Set("Client-Id", t.cfg.Twitch.ClientID)

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send guest star session request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return false, fmt.Errorf("guest star session request failed with status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read guest star session response: %w", err)
	}

	var sessionData struct {
		Data []struct {
			Guests []interface{} `json:"guests"`
		} `json:"data"`
	}

	err = json.Unmarshal(body, &sessionData)
	if err != nil {
		return false, fmt.Errorf("failed to parse guest star session response: %w", err)
	}

	if len(sessionData.Data) == 0 || len(sessionData.Data[0].Guests) == 0 {
		return false, nil
	}

	return true, nil
}

func (t *TwitchActions) IsGuestStarSessionActive(channel string) bool {
	return taskq.Compute(t.getQueue(channel), func() bool {
		botResp, err := t.helixClient.GetUsers(&helix.UsersParams{
			Logins: []string{util.BotUsername},
		})
		if err != nil || len(botResp.Data.Users) == 0 {
			slog.Error("Error getting bot user",
				slog.String("channel", channel),
				slog.Any("error", err),
			)
			return false
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
			return false
		}
		channelUser := channelResp.Data.Users[0]

		isSessionActive, err := t.getGuestStarSessionIsActive(channelUser.ID, botUser.ID)
		if err != nil {
			slog.Error("Error getting guest star session status",
				slog.String("channel", channel),
				slog.Any("error", err),
			)
			return false
		}

		return isSessionActive
	})
}

func (t *TwitchActions) SetEmoteMode(channel string, enabled bool) {
	t.getQueue(channel).Enqueue(func() {
		slog.Info("Set emote mode",
			slog.String("channel", channel),
			slog.Bool("enabled", enabled),
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

		setResp, err := t.helixClient.UpdateChatSettings(&helix.UpdateChatSettingsParams{
			BroadcasterID: channelUser.ID,
			ModeratorID:   botUser.ID,
			EmoteMode:     &enabled,
		})
		if err != nil {
			slog.Error("Error setting emote mode",
				slog.String("channel", channel),
				slog.Bool("enabled", enabled),
				slog.Any("error", err),
			)
			return
		}

		if setResp.StatusCode >= 400 {
			slog.Error("Set emote mode API error",
				slog.String("channel", channel),
				slog.Bool("enabled", enabled),
				slog.Any("error", setResp.Error),
				slog.Any("errorMsg", setResp.ErrorMessage),
			)
		}
	})
}

func (t *TwitchActions) GetViewerList(channel string) []string {
	result, err := t.ircClient.Userlist(channel)
	if err != nil {
		slog.Error("Error getting viewer list",
			slog.String("channel", channel),
			slog.Any("error", err),
		)

		return []string{}
	}

	if len(result) == 0 {
		return []string{}
	}

	return result
}

func (t *TwitchActions) DeleteMessage(channel, id string) {
	t.getQueue(channel).Enqueue(func() {
		slog.Info("Message has been deleted",
			slog.String("channel", channel),
			slog.String("message_id", id),
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

		banResp, err := t.helixClient.DeleteChatMessage(&helix.DeleteChatMessageParams{
			BroadcasterID: channelUser.ID,
			ModeratorID:   botUser.ID,
			MessageID:     id,
		})
		if err != nil {
			slog.Error("Error deleting message",
				slog.String("channel", channel),
				slog.String("message_id", id),
				slog.Any("error", err),
			)
			return
		}

		if banResp.StatusCode >= 400 {
			slog.Error("Delete message API error",
				slog.String("channel", channel),
				slog.String("message_id", id),
				slog.Any("error", err),
			)
		}
	})
}

func (t *TwitchActions) GetStartTime(channel string) time.Time {
	return taskq.Compute(t.getQueue(channel), func() time.Time {
		slog.Debug("Getting channel stream start time",
			slog.String("channel", channel),
		)

		res, err := t.helixClient.GetStreams(&helix.StreamsParams{
			UserLogins: []string{channel},
		})
		if err != nil {
			slog.Error("Failed to get stream info",
				slog.String("channel", channel),
				slog.Any("error", err),
			)
			return time.Time{}
		}

		for _, s := range res.Data.Streams {
			if strings.ToLower(s.UserLogin) == channel {
				return s.StartedAt
			}
		}

		return time.Time{}
	})
}

func (t *TwitchActions) GetViewerCount(channel string) int {
	return taskq.Compute(t.getQueue(channel), func() int {
		slog.Debug("Getting channel stream viewer count",
			slog.String("channel", channel),
		)

		res, err := t.helixClient.GetStreams(&helix.StreamsParams{
			UserLogins: []string{channel},
		})
		if err != nil {
			slog.Error("Failed to get stream info",
				slog.String("channel", channel),
				slog.Any("error", err),
			)
			return 0
		}

		for _, s := range res.Data.Streams {
			if strings.ToLower(s.UserLogin) == channel {
				return s.ViewerCount
			}
		}

		return 0
	})
}

func (t *TwitchActions) SendMessage(channel, text string) {
	t.getQueue(channel).Enqueue(func() {
		slog.Info("<<<",
			slog.String("channel", channel),
			slog.String("text", text),
		)

		t.ircClient.Say(channel, text)
	})
}

func (t *TwitchActions) SendForeignMessage(channel, text string) {
	t.getQueue(channel).Enqueue(func() {
		slog.Info("Send foreign message",
			slog.String("channel", channel),
			slog.String("text", text),
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

		sendMsgResp, err := t.helixClient.SendChatMessage(&helix.SendChatMessageParams{
			BroadcasterID: channelUser.ID,
			SenderID:      botUser.ID,
			Message:       text,
		})
		if err != nil {
			slog.Error("Error sending foreign message",
				slog.String("channel", channel),
				slog.String("message", text),
				slog.Any("error", err),
			)
			return
		}

		if sendMsgResp.StatusCode >= 400 {
			slog.Error("Send foreign message API error",
				slog.String("channel", channel),
				slog.String("message", text),
				slog.String("error", sendMsgResp.Error),
				slog.String("errorMsg", sendMsgResp.ErrorMessage),
			)
		}
	})
}

func (t *TwitchActions) TimeoutUser(channel, username string, duration time.Duration, reason string) {
	t.getQueue(channel).Enqueue(func() {
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
				slog.Any("error", banResp.Error),
				slog.Any("msg", banResp.ErrorMessage),
			)
		}
	})
}

func (t *TwitchActions) UnbanUser(channel, username string) {
	t.getQueue(channel).Enqueue(func() {
		slog.Info("User has been unbanned",
			slog.String("channel", channel),
			slog.String("username", username),
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
				slog.Any("error", unbanResp.Error),
				slog.Any("msg", unbanResp.ErrorMessage),
			)
		}
	})
}
