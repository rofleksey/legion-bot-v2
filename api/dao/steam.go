package dao

import "time"

type Comment struct {
	ID        string    `json:"id"`
	Author    string    `json:"author"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}
