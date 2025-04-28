package doctor

import (
	"fmt"
	"legion-bot-v2/chat"
	"legion-bot-v2/db"
	"legion-bot-v2/i18n"
	"legion-bot-v2/killer"
	"legion-bot-v2/timers"
	"legion-bot-v2/util"
	"log/slog"
	"math/rand/v2"
	"strings"
	"time"
	"unicode"
)

var _ killer.Killer = (*Doctor)(nil)

const (
	MadnessTimerName = "!!madness!!"
)

type Doctor struct {
	db.DB
	chat.Actions
	timers.Timers
	i18n.Localiser
}

func New(db db.DB, actions chat.Actions, timers timers.Timers, localiser i18n.Localiser) *Doctor {
	k := &Doctor{
		DB:        db,
		Actions:   actions,
		Timers:    timers,
		Localiser: localiser,
	}

	return k
}

func (d *Doctor) Name() string {
	return "doctor"
}

func (d *Doctor) Weight(channel string) int {
	chanState := d.GetState(channel)
	return chanState.Settings.Killers.Doctor.Weight
}

func (d *Doctor) Enabled(channel string) bool {
	chanState := d.GetState(channel)
	return chanState.Settings.Killers.Doctor.Enabled
}

func (d *Doctor) FixSettings(channel string) {
	chanState := d.GetState(channel)
	if chanState.Settings.Killers.Doctor == nil {
		d.UpdateState(chanState.Channel, func(chanState *db.ChannelState) {
			chanState.Settings.Killers.Doctor = db.DefaultDoctorSettings()
		})
	}
}

func (d *Doctor) startMadnessTimer(channel string) {
	d.StopTimer(channel, MadnessTimerName)

	chanState := d.GetState(channel)
	doctorSettings := chanState.Settings.Killers.Doctor
	lang := chanState.Settings.Language

	d.StartTimer(channel, MadnessTimerName, doctorSettings.Timeout, func() {
		d.UpdateState(channel, func(chanState *db.ChannelState) {
			chanState.Killer = ""
			chanState.KillerState = nil
			chanState.Date = time.Now()
			chanState.Stats["fail"]++
		})

		msg := d.GetLocalString(lang, "doctor_go_away", map[string]string{})
		d.SendMessage(channel, msg)
	})
}

func (d *Doctor) startMadness(channel string) {
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
		channelState.Killer = "doctor"
		channelState.Date = now
		channelState.Stats["total"]++
	})

	msg := d.GetLocalString(lang, "start_doctor", nil)
	d.SendMessage(channel, msg)

	d.startMadnessTimer(channel)

	slog.Info("Madness started", slog.String("channel", channel))
}

func (d *Doctor) Start(userMsg db.Message) {
	d.startMadness(userMsg.Channel)
}

func (d *Doctor) HandleMessage(userMsg db.Message) {
	chanState := d.GetState(userMsg.Channel)
	doctorSettings := chanState.Settings.Killers.Doctor
	now := time.Now()

	if chanState.Settings.Disabled {
		return
	}

	if d.handleCommands(userMsg) {
		return
	}

	user := chanState.UserMap[userMsg.Username]
	diff := now.Sub(chanState.Date)

	if diff < doctorSettings.MinDelayBetweenHits {
		return
	}

	if rand.Float64() > doctorSettings.ReactChance {
		return
	}

	if user.Health == "dead" || user.Health == "hooked" {
		return
	}

	if userMsg.Username == userMsg.Channel || userMsg.IsMod || strings.Contains(userMsg.Username, "bot") || userMsg.Username == util.BotOwner {
		return
	}

	d.handleHit(userMsg)
}

func (d *Doctor) handleCommands(userMsg db.Message) bool {
	chanState := d.GetState(userMsg.Channel)
	lang := chanState.Settings.Language

	switch {
	case strings.HasPrefix(userMsg.Text, "!killer"):
		msg := d.GetLocalString(lang, "commands_doctor", map[string]string{"STATS": fmt.Sprintf("https://leg.rofleksey.ru/#/stats/%s", userMsg.Channel)})
		d.SendMessage(userMsg.Channel, msg)
		return true
	}

	return false
}

func (d *Doctor) handleHit(userMsg db.Message) {
	d.UpdateState(userMsg.Channel, func(chanState *db.ChannelState) {
		chanState.Date = time.Now()
	})

	d.DeleteMessage(userMsg.Channel, userMsg.ID)
	d.SendMessage(userMsg.Channel, scrambleText(userMsg.Text))
}

func scrambleWord(word []rune) []rune {
	if len(word) <= 3 {
		return word
	}

	lastChar := word[len(word)-1]
	var punctuation []rune
	if !unicode.IsLetter(lastChar) && !unicode.IsNumber(lastChar) {
		punctuation = []rune{lastChar}
		word = word[:len(word)-1]
	}

	middle := word[1 : len(word)-1]
	rand.Shuffle(len(middle), func(i, j int) {
		middle[i], middle[j] = middle[j], middle[i]
	})

	scrambled := []rune{word[0]}
	scrambled = append(scrambled, middle...)
	scrambled = append(scrambled, word[len(word)-1])
	scrambled = append(scrambled, punctuation...)

	return scrambled
}

func scrambleText(text string) string {
	result := util.WordRegex.ReplaceAllStringFunc(text, func(match string) string {
		return string(scrambleWord([]rune(match)))
	})
	return result
}
