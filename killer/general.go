package killer

import (
	"legion-bot-v2/db"
)

type Killer interface {
	Name() string
	Enabled(channel string) bool
	FixSettings(channel string)
	Weight(channel string) int
	Start(userMsg db.Message)
	HandleMessage(userMsg db.Message)
}
