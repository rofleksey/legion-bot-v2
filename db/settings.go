package db

import "time"

type Settings struct {
	Disabled bool            `json:"disabled"`
	Language string          `json:"language"`
	Killers  KillersSettings `json:"killers"`
}

type KillersSettings struct {
	General *GeneralKillerSettings `json:"general"`
	Legion  *LegionSettings        `json:"legion"`
}

func DefaultSettings() Settings {
	return Settings{
		Disabled: false,
		Language: "en",
		Killers: KillersSettings{
			General: DefaultGeneralKillerSettings(),
			Legion:  DefaultLegionSettings(),
		},
	}
}

type GeneralKillerSettings struct {
	DelayBetweenKillers time.Duration `json:"delayBetweenKillers"`
}

func DefaultGeneralKillerSettings() *GeneralKillerSettings {
	return &GeneralKillerSettings{
		DelayBetweenKillers: 2 * time.Hour,
	}
}

type LegionSettings struct {
	Enabled                bool          `json:"enabled"`
	BodyBlockSuccessChance float64       `json:"bodyBlockSuccessChance"`
	DeepWoundTimeout       time.Duration `json:"deepWoundTimeout"`
	FatalHit               int           `json:"fatalHit"`
	FrenzyTimeout          time.Duration `json:"frenzyTimeout"`
	HitChance              float64       `json:"hitChance"`
	HookBanTime            time.Duration `json:"hookBanTime"`
	LockerGrabChance       float64       `json:"lockerGrabChance"`
	LockerStunChance       float64       `json:"lockerStunChance"`
	MinDelayBetweenHits    time.Duration `json:"minDelayBetweenHits"`
	PalletStunChance       float64       `json:"palletStunChance"`
	ReactChance            float64       `json:"reactChance"`
	BleedOutBanTime        time.Duration `json:"bleedOutBanTime"`
}

func DefaultLegionSettings() *LegionSettings {
	return &LegionSettings{
		Enabled:                true,
		BodyBlockSuccessChance: 0.2,
		DeepWoundTimeout:       time.Minute,
		FatalHit:               5,
		FrenzyTimeout:          3 * time.Minute,
		HitChance:              0.96,
		HookBanTime:            time.Minute,
		LockerGrabChance:       0.3,
		LockerStunChance:       0.25,
		MinDelayBetweenHits:    5 * time.Second,
		PalletStunChance:       0.18,
		ReactChance:            0.3,
		BleedOutBanTime:        30 * time.Second,
	}
}
