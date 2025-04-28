package db

type LegionState struct {
	HitCount int `json:"hitCount"`
}

type GhostFaceState struct {
	StalkedThisRound map[string]bool `json:"stalkedThisRound"`
}

type PinheadState struct {
	Word string `json:"word"`
}
