package db

import "time"

type User struct {
	Health string         `json:"health"`
	Marked bool           `json:"marked"`
	Stats  map[string]int `json:"stats"`
}

type ChannelState struct {
	Channel     string                  `json:"channel"`
	Killer      string                  `json:"killer"`
	KillerState any                     `json:"state"`
	Date        time.Time               `json:"date"`
	Stats       map[string]int          `json:"stats"`
	UserMap     map[string]*User        `json:"userMap"`
	Settings    Settings                `json:"settings"`
	GuestStar   ChannelGuestStarSession `json:"guestStar"`
	Subs        ChannelSubs             `json:"subs"`
}

type ChannelGuestStarSession struct {
	Active bool      `json:"active"`
	Date   time.Time `json:"date"`
}

type ChannelSubs struct {
	RaidID         string `json:"raidId"`
	GuestStarBegin string `json:"guestStarBegin"`
	GuestStarEnd   string `json:"guestStarEnd"`
}

type Message struct {
	ID       string
	Channel  string
	Username string
	IsMod    bool
	Text     string
}

type PartialMessage struct {
	Channel  string
	Username string
	Text     string
}
