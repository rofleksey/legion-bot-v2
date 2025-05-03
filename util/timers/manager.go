package timers

import (
	"sync"
	"time"
)

type TimerInfo struct {
	timer    *time.Timer
	deadline time.Time
}

type Manager struct {
	timers map[string]map[string]*TimerInfo
	mu     sync.Mutex
}

func NewManager() Timers {
	return &Manager{
		timers: make(map[string]map[string]*TimerInfo),
	}
}

func (tm *Manager) StartTimer(channel, name string, duration time.Duration, callback func()) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.stopTimerUnsafe(channel, name)

	if _, exists := tm.timers[channel]; !exists {
		tm.timers[channel] = make(map[string]*TimerInfo)
	}

	deadline := time.Now().Add(duration)
	tm.timers[channel][name] = &TimerInfo{
		timer: time.AfterFunc(duration, func() {
			callback()

			tm.mu.Lock()
			defer tm.mu.Unlock()

			delete(tm.timers[channel], name)
		}),
		deadline: deadline,
	}
}

func (tm *Manager) StopTimer(channel, name string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.stopTimerUnsafe(channel, name)
}

func (tm *Manager) stopTimerUnsafe(channel, name string) {
	if channelTimers, exists := tm.timers[channel]; exists {
		if timerInfo, exists := channelTimers[name]; exists {
			timerInfo.timer.Stop()
			delete(channelTimers, name)
		}
	}
}

func (tm *Manager) StopChannelTimers(channel string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	channelTimers := tm.timers[channel]
	for name, timerInfo := range channelTimers {
		timerInfo.timer.Stop()
		delete(channelTimers, name)
	}

	delete(tm.timers, channel)
}

func (tm *Manager) GetRemainingTime(channel, name string) time.Duration {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if channelTimers, exists := tm.timers[channel]; exists {
		if timerInfo, exists := channelTimers[name]; exists {
			return time.Until(timerInfo.deadline)
		}
	}
	return 0
}
