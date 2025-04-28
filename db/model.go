package db

import "time"

type Channel struct {
	Name string
	Lang string
}

type User struct {
	Health string         `json:"health"`
	Marked bool           `json:"marked"`
	Stats  map[string]int `json:"stats"`
}

type ChannelState struct {
	Channel     string           `json:"channel"`
	Killer      string           `json:"killer"`
	KillerState any              `json:"state"`
	Date        time.Time        `json:"date"`
	Stats       map[string]int   `json:"stats"`
	UserMap     map[string]*User `json:"userMap"`
	Settings    Settings         `json:"settings"`
}

type Message struct {
	Channel  string
	Username string
	IsMod    bool
	Text     string
}
