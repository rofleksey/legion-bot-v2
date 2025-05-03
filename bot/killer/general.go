package killer

import (
	"legion-bot-v2/db"
	"time"
)

type Killer interface {
	Name() string
	Enabled(channel string) bool
	FixSettings(chanState *db.ChannelState) bool
	Weight(channel string) int
	Start(userMsg db.Message)
	HandleMessage(userMsg db.Message)
	HandleWhisper(userMsg db.PartialMessage)
	TimeRemaining(channel string) time.Duration
}
