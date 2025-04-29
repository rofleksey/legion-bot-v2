package common

import "context"

type DetectedUser struct {
	Username string `json:"username"`
	Site     string `json:"site"`
}

type Detector interface {
	Detect(ctx context.Context, username string) ([]DetectedUser, error)
	Name() string
}
