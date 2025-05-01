package chat

import (
	"legion-bot-v2/util"
	"log/slog"
	"time"
)

var _ Actions = (*ConsoleActions)(nil)

type ConsoleActions struct {
}

func (a *ConsoleActions) SetEmoteMode(channel string, enabled bool) {
	slog.Debug("Set emote mode",
		slog.String("channel", channel),
		slog.Bool("enabled", enabled),
	)
}

func (a *ConsoleActions) GetViewerList(channel string) []string {
	return []string{util.BotUsername}
}

func (a *ConsoleActions) DeleteMessage(channel, id string) {
	slog.Debug("Deleting chat message",
		slog.String("channel", channel),
		slog.String("message_id", id),
	)
}

func (a *ConsoleActions) GetViewerCount(channel string) int {
	slog.Debug("Getting channel stream viewer count",
		slog.String("channel", channel),
	)

	return 15
}

func (a *ConsoleActions) GetStartTime(channel string) time.Time {
	slog.Debug("Getting channel stream start time",
		slog.String("channel", channel),
	)

	return time.Now().Add(-2 * time.Hour)
}

func (a *ConsoleActions) SendMessage(channel, text string) {
	slog.Debug("<<<",
		slog.String("channel", channel),
		slog.String("text", text),
	)
}

func (a *ConsoleActions) SendForeignMessage(channel, text string) {
	slog.Debug("Send foreign message",
		slog.String("channel", channel),
		slog.String("text", text),
	)
}

func (a *ConsoleActions) TimeoutUser(channel, username string, duration time.Duration, reason string) {
	slog.Debug("User has been timeout",
		slog.String("channel", channel),
		slog.String("username", username),
		slog.Duration("duration", duration),
		slog.String("reason", reason),
	)
}

func (a *ConsoleActions) UnbanUser(channel, username string) {
	slog.Debug("User has been unbanned",
		slog.String("channel", channel),
		slog.String("username", username),
	)
}
