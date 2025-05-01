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

type DredgeState struct {
	Votes map[string]string `json:"votes"`
}
