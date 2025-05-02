package chat

import (
	"github.com/gempir/go-twitch-irc/v4"
	"github.com/nicklaw5/helix/v2"
	"legion-bot-v2/config"
	"legion-bot-v2/taskq"
	"legion-bot-v2/util"
	"log/slog"
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

func (t *TwitchActions) getIds(channel string) (string, string) {
	botResp, err := t.helixClient.GetUsers(&helix.UsersParams{
		Logins: []string{util.BotUsername},
	})
	if err != nil {
		slog.Error("Error getting bot user",
			slog.String("channel", channel),
			slog.Any("error", err),
		)
		return "", ""
	}
	if botResp.StatusCode >= 400 || len(botResp.Data.Users) == 0 {
		slog.Error("Error getting bot user",
			slog.String("channel", channel),
			slog.String("error", botResp.Error),
			slog.String("errorMsg", botResp.ErrorMessage),
		)
		return "", ""
	}
	botUser := botResp.Data.Users[0]

	channelResp, err := t.helixClient.GetUsers(&helix.UsersParams{
		Logins: []string{channel},
	})
	if err != nil {
		slog.Error("Error getting channel user",
			slog.String("channel", channel),
			slog.Any("error", err),
		)
		return "", ""
	}
	if channelResp.StatusCode >= 400 || len(channelResp.Data.Users) == 0 {
		slog.Error("Error getting channel user",
			slog.String("channel", channel),
			slog.String("error", channelResp.Error),
			slog.String("errorMsg", channelResp.ErrorMessage),
		)
		return "", ""
	}
	channelUser := channelResp.Data.Users[0]

	return channelUser.ID, botUser.ID
}

func (t *TwitchActions) SetEmoteMode(channel string, enabled bool) {
	t.getQueue(channel).Enqueue(func() {
		slog.Info("Set emote mode",
			slog.String("channel", channel),
			slog.Bool("enabled", enabled),
		)

		channelUserID, botUserID := t.getIds(channel)
		if channelUserID == "" {
			return
		}

		setResp, err := t.helixClient.UpdateChatSettings(&helix.UpdateChatSettingsParams{
			BroadcasterID: channelUserID,
			ModeratorID:   botUserID,
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

		channelUserID, botUserID := t.getIds(channel)
		if channelUserID == "" {
			return
		}

		banResp, err := t.helixClient.DeleteChatMessage(&helix.DeleteChatMessageParams{
			BroadcasterID: channelUserID,
			ModeratorID:   botUserID,
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
		if res.StatusCode >= 400 {
			slog.Error("Failed to get stream info",
				slog.String("channel", channel),
				slog.String("error", res.Error),
				slog.String("errorMsg", res.ErrorMessage),
			)
			return time.Time{}
		}

		for _, s := range res.Data.Streams {
			if strings.ToLower(s.UserLogin) == channel {
				return s.StartedAt
			}
		}

		slog.Warn("Stream start time not found",
			slog.String("channel", channel),
		)

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
		if res.StatusCode >= 400 {
			slog.Error("Failed to get stream info",
				slog.String("channel", channel),
				slog.String("error", res.Error),
				slog.String("errorMsg", res.ErrorMessage),
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

		channelUserID, botUserID := t.getIds(channel)
		if channelUserID == "" {
			return
		}

		sendMsgResp, err := t.helixClient.SendChatMessage(&helix.SendChatMessageParams{
			BroadcasterID: channelUserID,
			SenderID:      botUserID,
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

		channelUserID, botUserID := t.getIds(channel)
		if channelUserID == "" {
			return
		}

		userResp, err := t.helixClient.GetUsers(&helix.UsersParams{
			Logins: []string{username},
		})
		if err != nil {
			slog.Error("Error getting user to ban",
				slog.String("channel", channel),
				slog.Any("error", err),
			)
			return
		}
		if userResp.StatusCode >= 400 || len(userResp.Data.Users) == 0 {
			slog.Error("Error getting user to ban",
				slog.String("channel", channel),
				slog.String("error", userResp.Error),
				slog.String("errorMsg", userResp.ErrorMessage),
			)
			return
		}
		banUser := userResp.Data.Users[0]

		banResp, err := t.helixClient.BanUser(&helix.BanUserParams{
			BroadcasterID: channelUserID,
			ModeratorId:   botUserID,
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

		channelUserID, botUserID := t.getIds(channel)
		if channelUserID == "" {
			return
		}

		userResp, err := t.helixClient.GetUsers(&helix.UsersParams{
			Logins: []string{username},
		})
		if err != nil {
			slog.Error("Error getting user to unban",
				slog.String("channel", channel),
				slog.Any("error", err),
			)
			return
		}
		if userResp.StatusCode >= 400 || len(userResp.Data.Users) == 0 {
			slog.Error("Error getting user to unban",
				slog.String("channel", channel),
				slog.String("error", userResp.Error),
				slog.String("errorMsg", userResp.ErrorMessage),
			)
			return
		}
		banUser := userResp.Data.Users[0]

		unbanResp, err := t.helixClient.UnbanUser(&helix.UnbanUserParams{
			BroadcasterID: channelUserID,
			ModeratorID:   botUserID,
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
