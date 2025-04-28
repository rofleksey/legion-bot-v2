package chat

import (
	"time"
)

type Actions interface {
	DeleteMessage(channel, id string)
	SendMessage(channel, text string)
	TimeoutUser(channel, username string, duration time.Duration, reason string)
	GetStartTime(channel string) time.Time
	GetViewerCount(channel string) int
	UnbanUser(channel, username string)
}
