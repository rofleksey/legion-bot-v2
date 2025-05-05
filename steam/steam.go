package steam

import (
	"context"
	"fmt"
	"github.com/samber/do"
	"legion-bot-v2/api/dao"
	"legion-bot-v2/bot"
	"legion-bot-v2/config"
	"legion-bot-v2/db"
	"legion-bot-v2/steam/steam_api"
	"legion-bot-v2/twitch/chat"
	"log/slog"
	"strings"
	"time"
)

const checkInterval = 30 * time.Minute

type Steam interface {
	Run(ctx context.Context)
	UpdatePinnedComment(channel string)
}

type Client struct {
	botInstance *bot.Bot
	*steam_api.Client
	db.DB
	chat.Actions
}

func NewClient(di *do.Injector) (Steam, error) {
	cfg := do.MustInvoke[*config.Config](di)

	client, err := steam_api.NewClient(cfg.Steam.SessionID, cfg.Steam.SteamLoginSecure)
	if err != nil {
		return nil, fmt.Errorf("steam_api.NewClient: %w", err)
	}

	return &Client{
		Client:      client,
		botInstance: do.MustInvoke[*bot.Bot](di),
		DB:          do.MustInvoke[db.DB](di),
		Actions:     do.MustInvoke[chat.Actions](di),
	}, nil
}

func (c *Client) Run(ctx context.Context) {
	ticker := time.NewTicker(checkInterval)

	c.doPeriodicJobs()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.doPeriodicJobs()
		}
	}
}

func (c *Client) UpdatePinnedComment(channel string) {
	chanState := c.GetState(channel)
	steamSettings := chanState.Settings.Steam
	steamState := chanState.Steam
	steamId64 := strings.TrimSpace(steamSettings.SteamID64)
	pinnedCommentText := strings.TrimSpace(steamSettings.PinnedCommentText)

	if chanState.Settings.Disabled ||
		time.Now().Before(chanState.UserTimeout) ||
		steamId64 == "" {
		return
	}

	if steamState.PinnedCommentID != "" {
		if err := c.DeleteComment(steamId64, steamState.PinnedCommentID); err != nil {
			slog.Error("Failed to delete pinned comment",
				slog.String("channel", channel),
				slog.String("steamId64", steamId64),
				slog.Any("error", err),
			)
			return
		}

		c.UpdateState(channel, func(state *db.ChannelState) {
			state.Steam.PinnedCommentID = ""
		})

		slog.Debug("Pinned comment deleted",
			slog.String("channel", channel),
			slog.String("steamId64", steamId64),
		)
	}

	if pinnedCommentText != "" {
		newId, err := c.PostComment(steamId64, pinnedCommentText)
		if err != nil {
			slog.Error("Failed to post comment",
				slog.String("channel", channel),
				slog.String("steamId64", steamId64),
				slog.Any("error", err),
			)
			return
		}

		c.UpdateState(channel, func(state *db.ChannelState) {
			state.Steam.PinnedCommentID = newId
		})

		slog.Info("Pinned comment updated",
			slog.String("channel", channel),
			slog.String("steamId64", steamId64),
		)
	}
}

func (c *Client) handleComments(channel string) {
	slog.Debug("Starting handleComments",
		slog.String("channel", channel),
	)

	chanState := c.GetState(channel)
	steamSettings := chanState.Settings.Steam
	steamState := chanState.Steam
	steamId64 := strings.TrimSpace(steamSettings.SteamID64)

	curComments, err := c.GetLatestComments(steamId64)
	if err != nil {
		slog.Error("Error getting latest steam comments",
			slog.String("channel", channel),
			slog.String("steamId64", steamId64),
			slog.Any("error", err),
		)
		return
	}

	if len(curComments) == 0 {
		slog.Debug("No comments found",
			slog.String("channel", channel),
		)
		c.UpdatePinnedComment(channel)
		return
	}

	var lastComment dao.Comment

	for _, curComment := range curComments {
		if curComment.Timestamp.After(lastComment.Timestamp) {
			lastComment = curComment
		}
	}

	if lastComment.Timestamp.After(steamState.LastCommentTime) && lastComment.ID != steamState.PinnedCommentID {
		slog.Debug("Got a new comment on steam",
			slog.String("channel", channel),
		)

		if !steamState.LastCommentTime.IsZero() {
			c.botInstance.HandleNewSteamComment(channel, lastComment)
		}

		c.UpdateState(channel, func(chanState *db.ChannelState) {
			chanState.Steam.LastCommentTime = lastComment.Timestamp
		})

		c.UpdatePinnedComment(channel)
	} else {
		slog.Debug("Comments not changed since last time",
			slog.String("channel", channel),
		)
	}
}

func (c *Client) doPeriodicJobs() {
	slog.Debug("Starting steam periodic jobs")
	defer slog.Debug("Steam periodic jobs finished")

	channels := c.GetAllChannelNames()

	for _, channel := range channels {
		slog.Debug("Processing steam jobs for channel",
			slog.String("channel", channel),
		)

		chanState := c.GetState(channel)
		steamSettings := chanState.Settings.Steam
		steamId64 := strings.TrimSpace(steamSettings.SteamID64)

		if chanState.Settings.Disabled || time.Now().Before(chanState.UserTimeout) {
			slog.Debug("Bot is disabled, skipping steam stuff",
				slog.String("channel", channel),
			)
			continue
		}

		if steamId64 == "" {
			slog.Debug("No steam id found in steam settings",
				slog.String("channel", channel),
			)
			continue
		}

		if steamSettings.NotifyNewComments || strings.TrimSpace(steamSettings.PinnedCommentText) != "" {
			c.handleComments(channel)
		}
	}
}
