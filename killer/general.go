package killer

import (
	"legion-bot-v2/db"
)

type Killer interface {
	Start(userMsg db.Message)
	HandleMessage(userMsg db.Message)
}
