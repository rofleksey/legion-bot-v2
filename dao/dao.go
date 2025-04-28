package dao

import "legion-bot-v2/db"

type ImportRequest struct {
	Legions []LegionData `json:"legions"`
}

type LegionData struct {
	Channel  string              `json:"channel"`
	State    string              `json:"state"`
	Date     int64               `json:"date"`
	HitCount int                 `json:"hitCount"`
	UserMap  map[string]*db.User `json:"userMap"`
	Stats    map[string]int      `json:"stats"`
	Settings db.Settings         `json:"settings"`
}

type TwitchUser struct {
	ID              string `json:"id"`
	Login           string `json:"login"`
	DisplayName     string `json:"displayName"`
	ProfileImageURL string `json:"profileImageUrl"`
	Email           string `json:"email"`
}
