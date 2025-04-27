package chat

import (
	"time"
)

type Actions interface {
	SendMessage(channel, text string)
	TimeoutUser(channel, username string, duration time.Duration, reason string)
	UnbanUser(channel, username string)
}
