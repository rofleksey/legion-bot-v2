package db

import "time"

func NewUser() *User {
	return &User{
		Health: "healthy",
		Stats:  make(map[string]int),
	}
}

func NewChannelState(channel string) ChannelState {
	return ChannelState{
		Channel: channel,
		Date:    time.Time{},
		Stats: map[string]int{
			"total":     0,
			"success":   0,
			"fail":      0,
			"miss":      0,
			"hits":      0,
			"bleedOuts": 0,
			"stuns":     0,
			"bodyBlock": 0,
		},
		UserMap:  make(map[string]*User),
		Settings: DefaultSettings(),
	}
}
