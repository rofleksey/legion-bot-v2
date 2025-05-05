package chat

import (
	"time"
)

type Actions interface {
	GetUserIDByUsername(username string) string
	DeleteMessage(channel, id string)
	SendMessage(channel, text string)
	SendForeignMessage(channel, text string)
	TimeoutUser(channel, username string, duration time.Duration, reason string)
	GetStartTime(channel string) time.Time
	GetViewerCount(channel string) int
	UnbanUser(channel, username string)
	GetViewerList(channel string) []string
	SetEmoteMode(channel string, enabled bool)
	Shutdown()
}
