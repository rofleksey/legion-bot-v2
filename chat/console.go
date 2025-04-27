package chat

import (
	"log/slog"
	"time"
)

var _ Actions = (*ConsoleActions)(nil)

type ConsoleActions struct {
}

func (n ConsoleActions) SendMessage(channel, text string) {
	slog.Info("<<<",
		slog.String("channel", channel),
		slog.String("text", text),
	)
}

func (n ConsoleActions) TimeoutUser(channel, username string, duration time.Duration, reason string) {
	slog.Info("User has been timeout",
		slog.String("channel", channel),
		slog.String("username", username),
		slog.Duration("duration", duration),
		slog.String("reason", reason),
	)
}

func (n ConsoleActions) UnbanUser(channel, username string) {
	slog.Info("User has been unbanned",
		slog.String("channel", channel),
		slog.String("username", username),
	)
}
