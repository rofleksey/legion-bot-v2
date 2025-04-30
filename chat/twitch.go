package chat

import (
	"github.com/gempir/go-twitch-irc/v4"
	"github.com/nicklaw5/helix/v2"
	"legion-bot-v2/util"
	"log/slog"
	"strings"
	"sync"
	"time"
)

var _ Actions = (*TwitchActions)(nil)

type TwitchActions struct {
	ircClient   *twitch.Client
	helixClient *helix.Client
	queues      map[string]*util.TaskQueue
	mu          sync.Mutex
}

func NewTwitchActions(ircClient *twitch.Client, helixClient *helix.Client) *TwitchActions {
	return &TwitchActions{
		ircClient:   ircClient,
		helixClient: helixClient,
		queues:      make(map[string]*util.TaskQueue),
	}
}

func (t *TwitchActions) getQueue(channel string) *util.TaskQueue {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.queues[channel]; !exists {
		t.queues[channel] = util.NewTaskQueue(1, 1, 1)
	}
	return t.queues[channel]
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

		t.helixClient.SetUserAccessToken(botUser.ID)

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
}

func (t *TwitchActions) GetViewerCount(channel string) int {
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

		t.helixClient.SetUserAccessToken(botUser.ID)

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

		t.helixClient.SetUserAccessToken(botUser.ID)

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
