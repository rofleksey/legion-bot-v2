package dao

import (
	"encoding/json"
	"github.com/nicklaw5/helix/v2"
)

type EventSubNotification struct {
	Subscription helix.EventSubSubscription `json:"subscription"`
	Challenge    string                     `json:"challenge"`
	Event        json.RawMessage            `json:"event"`
}

type EventSubGuestStarEvent struct {
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
}
