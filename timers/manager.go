package timers

import (
	"sync"
	"time"
)

type Manager struct {
	timers map[string]map[string]*time.Timer
	mu     sync.Mutex
}

func NewManager() *Manager {
	return &Manager{
		timers: make(map[string]map[string]*time.Timer),
	}
}

func (tm *Manager) StartTimer(channel, name string, duration time.Duration, callback func()) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.stopTimerUnsafe(channel, name)

	if _, exists := tm.timers[channel]; !exists {
		tm.timers[channel] = make(map[string]*time.Timer)
	}

	tm.timers[channel][name] = time.AfterFunc(duration, func() {
		callback()

		tm.mu.Lock()
		defer tm.mu.Unlock()

		delete(tm.timers[channel], name)
	})
}

func (tm *Manager) StopTimer(channel, name string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.stopTimerUnsafe(channel, name)
}

func (tm *Manager) stopTimerUnsafe(channel, name string) {
	if channelTimers, exists := tm.timers[channel]; exists {
		if timer, exists := channelTimers[name]; exists {
			timer.Stop()

			delete(channelTimers, name)
		}
	}
}

func (tm *Manager) StopChannelTimers(channel string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	channelTimers := tm.timers[channel]
	for name, timer := range channelTimers {
		timer.Stop()
		delete(channelTimers, name)
	}

	delete(tm.timers, channel)
}
