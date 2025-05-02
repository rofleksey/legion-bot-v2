package chat

import (
	"github.com/gempir/go-twitch-irc/v4"
	"github.com/jellydator/ttlcache/v3"
	"github.com/nicklaw5/helix/v2"
	"github.com/samber/do"
	"legion-bot-v2/config"
	"legion-bot-v2/util"
	"legion-bot-v2/util/taskq"
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

	queueMutex sync.Mutex
	queues     map[string]*taskq.Queue

	userIdCache *ttlcache.Cache[string, string] // username -> userId
}

func NewTwitchActions(di *do.Injector) Actions {
	userIdCache := ttlcache.New(
		ttlcache.WithTTL[string, string](24 * time.Hour),
	)

	go userIdCache.Start()

	return &TwitchActions{
		cfg:         do.MustInvoke[*config.Config](di),
		accessToken: do.MustInvokeNamed[string](di, "userAccessToken"),
		ircClient:   do.MustInvoke[*twitch.Client](di),
		helixClient: do.MustInvokeNamed[*helix.Client](di, "helixClient"),
		queues:      make(map[string]*taskq.Queue),
		userIdCache: userIdCache,
	}
}

func (t *TwitchActions) getQueue(channel string) *taskq.Queue {
	t.queueMutex.Lock()
	defer t.queueMutex.Unlock()

	if _, exists := t.queues[channel]; !exists {
		t.queues[channel] = taskq.New(1, 1, 1)
	}
	return t.queues[channel]
}

func (t *TwitchActions) Shutdown() {
	t.queueMutex.Lock()
	defer t.queueMutex.Unlock()

	for _, q := range t.queues {
		q.Shutdown()
	}
}

func (t *TwitchActions) GetUserIDByUsername(username string) string {
	cacheItem := t.userIdCache.Get(username)
	if cacheItem != nil {
		return cacheItem.Value()
	}

	res, err := t.helixClient.GetUsers(&helix.UsersParams{
		Logins: []string{username},
	})
	if err != nil {
		slog.Error("Error getting user",
			slog.String("username", username),
			slog.Any("error", err),
		)
		return ""
	}
	if res.StatusCode >= 400 || len(res.Data.Users) == 0 {
		slog.Error("Error getting user",
			slog.String("username", username),
			slog.String("error", res.Error),
			slog.String("errorMsg", res.ErrorMessage),
		)
		return ""
	}
	user := res.Data.Users[0]

	t.userIdCache.Set(username, user.ID, ttlcache.DefaultTTL)

	return user.ID
}

func (t *TwitchActions) SetEmoteMode(channel string, enabled bool) {
	t.getQueue(channel).Enqueue(func() {
		slog.Info("Set emote mode",
			slog.String("channel", channel),
			slog.Bool("enabled", enabled),
		)

		channelUserID := t.GetUserIDByUsername(channel)
		if channelUserID == "" {
			return
		}

		setResp, err := t.helixClient.UpdateChatSettings(&helix.UpdateChatSettingsParams{
			BroadcasterID: channelUserID,
			ModeratorID:   util.BotUserID,
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

		channelUserID := t.GetUserIDByUsername(channel)
		if channelUserID == "" {
			return
		}

		banResp, err := t.helixClient.DeleteChatMessage(&helix.DeleteChatMessageParams{
			BroadcasterID: channelUserID,
			ModeratorID:   util.BotUserID,
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

		channelUserID := t.GetUserIDByUsername(channel)
		if channelUserID == "" {
			return
		}

		sendMsgResp, err := t.helixClient.SendChatMessage(&helix.SendChatMessageParams{
			BroadcasterID: channelUserID,
			SenderID:      util.BotUserID,
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

		channelUserID := t.GetUserIDByUsername(channel)
		if channelUserID == "" {
			return
		}

		banUserID := t.GetUserIDByUsername(username)
		if banUserID == "" {
			return
		}

		banResp, err := t.helixClient.BanUser(&helix.BanUserParams{
			BroadcasterID: channelUserID,
			ModeratorId:   util.BotUserID,
			Body: helix.BanUserRequestBody{
				Duration: int(duration.Seconds()),
				Reason:   reason,
				UserId:   banUserID,
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

		channelUserID := t.GetUserIDByUsername(channel)
		if channelUserID == "" {
			return
		}

		unbanUserID := t.GetUserIDByUsername(username)
		if unbanUserID == "" {
			return
		}

		unbanResp, err := t.helixClient.UnbanUser(&helix.UnbanUserParams{
			BroadcasterID: channelUserID,
			ModeratorID:   util.BotUserID,
			UserID:        unbanUserID,
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
