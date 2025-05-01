package dredge

import (
	"github.com/mitchellh/mapstructure"
	"legion-bot-v2/chat"
	"legion-bot-v2/db"
	"legion-bot-v2/gpt"
	"legion-bot-v2/i18n"
	"legion-bot-v2/killer"
	"legion-bot-v2/timers"
	"legion-bot-v2/util"
	"log/slog"
	"strings"
	"time"
)

var _ killer.Killer = (*Dredge)(nil)

const (
	NightfallTimer = "!!nightfall!!"
)

type Dredge struct {
	db.DB
	chat.Actions
	timers.Timers
	i18n.Localiser
	gpt.Gpt
}

func New(db db.DB, actions chat.Actions, timers timers.Timers, localiser i18n.Localiser, g gpt.Gpt) *Dredge {
	k := &Dredge{
		DB:        db,
		Actions:   actions,
		Timers:    timers,
		Localiser: localiser,
		Gpt:       g,
	}

	return k
}

func (d *Dredge) Name() string {
	return "dredge"
}

func (d *Dredge) Weight(channel string) int {
	chanState := d.GetState(channel)
	return chanState.Settings.Killers.Dredge.Weight
}

func (d *Dredge) Enabled(channel string) bool {
	chanState := d.GetState(channel)
	return chanState.Settings.Killers.Dredge.Enabled
}

func (d *Dredge) FixSettings(chanState *db.ChannelState) bool {
	if chanState.Settings.Killers.Dredge != nil {
		return false
	}

	chanState.Settings.Killers.Dredge = db.DefaultDredgeSettings()

	return true
}

func (d *Dredge) startNightfallTimer(channel string) {
	d.StopTimer(channel, NightfallTimer)

	chanState := d.GetState(channel)
	dredgeSettings := chanState.Settings.Killers.Dredge

	d.StartTimer(channel, NightfallTimer, dredgeSettings.Timeout, func() {
		d.onNightfallEnd(channel)
	})
}

func (d *Dredge) onNightfallEnd(channel string) {
	chanState := d.GetState(channel)
	dredgeSettings := chanState.Settings.Killers.Dredge
	lang := chanState.Settings.Language

	d.SetEmoteMode(channel, false)

	var dredgeState db.DredgeState
	if err := mapstructure.Decode(chanState.KillerState, &dredgeState); err != nil {
		slog.Error("Failed to decode killer state",
			slog.String("channel", channel),
			slog.Any("error", err),
		)
	}

	var counters map[string]int
	for _, otherUsername := range dredgeState.Votes {
		counters[otherUsername]++
	}

	var maxCounter int
	for _, count := range counters {
		if count > maxCounter {
			maxCounter = count
		}
	}

	var usernamesToHook []string
	for username, count := range counters {
		if count != maxCounter {
			continue
		}

		usernamesToHook = append(usernamesToHook, username)
	}

	if maxCounter <= 1 || len(usernamesToHook) != 1 {
		d.UpdateState(channel, func(chanState *db.ChannelState) {
			chanState.Killer = ""
			chanState.KillerState = nil
			chanState.Date = time.Now()
			chanState.Stats["fail"]++
		})

		msg := d.GetLocalString(lang, "dredge_go_away", map[string]string{})
		d.SendMessage(channel, msg)
		return
	}

	username := usernamesToHook[0]

	d.UpdateState(channel, func(chanState *db.ChannelState) {
		chanState.Killer = ""
		chanState.KillerState = nil
		chanState.Date = time.Now()
		chanState.UserMap[username].Health = "hooked"
		chanState.UserMap[username].Stats["hooks"]++
		chanState.Stats["success"]++
	})

	d.TimeoutUser(channel, username, dredgeSettings.HookBanTime, "")

	msg := d.GetLocalString(lang, "dredge_hit_dead", map[string]string{"USERNAME": username})
	d.SendMessage(channel, msg)
}

func (d *Dredge) startNightfall(channel string) {
	startState := d.GetState(channel)
	lang := startState.Settings.Language
	now := time.Now()

	if startState.Killer != "" {
		slog.Warn("Killer is already summoned",
			slog.String("channel", channel),
		)
		return
	}

	d.UpdateState(channel, func(channelState *db.ChannelState) {
		channelState.Killer = "dredge"
		channelState.KillerState = db.DredgeState{
			Votes: make(map[string]string),
		}
		channelState.Date = now
		channelState.Stats["total"]++
	})

	msg := d.GetLocalString(lang, "start_dredge", nil)
	d.SendMessage(channel, msg)

	d.SetEmoteMode(channel, true)
	d.startNightfallTimer(channel)

	slog.Info("Nightfall started", slog.String("channel", channel))
}

func (d *Dredge) Start(userMsg db.Message) {
	d.startNightfall(userMsg.Channel)
}

func (d *Dredge) HandleMessage(_ db.Message) {

}

func (d *Dredge) HandleWhisper(userMsg db.PartialMessage) {
	chanState := d.GetState(userMsg.Channel)

	if chanState.Settings.Disabled {
		return
	}

	if userMsg.Username == userMsg.Channel || strings.Contains(userMsg.Username, "bot") || userMsg.Username == util.BotOwner {
		return
	}

	user, userExists := chanState.UserMap[userMsg.Username]
	if !userExists {
		user = db.NewUser()
		d.UpdateState(userMsg.Channel, func(chanState *db.ChannelState) {
			chanState.UserMap[userMsg.Username] = user
		})
	}

	var dredgeState db.DredgeState
	if err := mapstructure.Decode(chanState.KillerState, &dredgeState); err != nil {
		slog.Error("Failed to decode killer state",
			slog.String("channel", userMsg.Channel),
			slog.Any("error", err),
		)
	}

	otherUsername := strings.ToLower(strings.TrimSpace(strings.ReplaceAll(userMsg.Text, "@", "")))
	dredgeState.Votes[userMsg.Username] = otherUsername

	d.UpdateState(userMsg.Channel, func(chanState *db.ChannelState) {
		chanState.KillerState = dredgeState
	})
}
