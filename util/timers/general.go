package timers

import "time"

type Timers interface {
	StartTimer(channel, name string, duration time.Duration, callback func())
	StopTimer(channel, name string)
	StopChannelTimers(channel string)
}
