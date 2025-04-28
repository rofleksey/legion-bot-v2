package pinhead

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"legion-bot-v2/chat"
	"legion-bot-v2/db"
	"legion-bot-v2/gpt"
	"legion-bot-v2/i18n"
	"legion-bot-v2/killer"
	"legion-bot-v2/timers"
	"log/slog"
	"strings"
	"time"
)

var _ killer.Killer = (*Pinhead)(nil)

const (
	BoxTimerName = "!!box!!"
)

type Pinhead struct {
	db.DB
	chat.Actions
	timers.Timers
	i18n.Localiser
	gpt.Gpt
}

func New(db db.DB, actions chat.Actions, timers timers.Timers, localiser i18n.Localiser, g gpt.Gpt) *Pinhead {
	k := &Pinhead{
		DB:        db,
		Actions:   actions,
		Timers:    timers,
		Localiser: localiser,
		Gpt:       g,
	}

	return k
}

func (p *Pinhead) Name() string {
	return "pinhead"
}

func (p *Pinhead) Weight(channel string) int {
	chanState := p.GetState(channel)
	return chanState.Settings.Killers.Pinhead.Weight
}

func (p *Pinhead) Enabled(channel string) bool {
	chanState := p.GetState(channel)
	return chanState.Settings.Killers.Pinhead.Enabled
}

func (p *Pinhead) FixSettings(channel string) {
	chanState := p.GetState(channel)
	if chanState.Settings.Killers.Pinhead == nil {
		p.UpdateState(chanState.Channel, func(chanState *db.ChannelState) {
			chanState.Settings.Killers.Pinhead = db.DefaultPinheadSettings()
		})
	}
}

func (p *Pinhead) startRecoverTimer(channel, username string) {
	p.StopTimer(channel, username)

	chanState := p.GetState(channel)
	pinheadSettings := chanState.Settings.Killers.Pinhead

	p.StartTimer(channel, username, pinheadSettings.BleedOutBanTime, func() {
		p.UpdateState(channel, func(chanState *db.ChannelState) {
			chanState.UserMap[username].Health = "injured"
		})
	})
}

func (p *Pinhead) startDeadTimer(channel, username string) {
	p.StopTimer(channel, username)

	chanState := p.GetState(channel)
	pinheadSettings := chanState.Settings.Killers.Pinhead
	lang := chanState.Settings.Language

	p.StartTimer(channel, username, pinheadSettings.DeepWoundTimeout, func() {
		p.UpdateState(channel, func(chanState *db.ChannelState) {
			chanState.Stats["bleedOuts"]++

			chanState.UserMap[username].Health = "dead"
			chanState.Stats["bleedOuts"]++
		})

		p.TimeoutUser(channel, username, pinheadSettings.BleedOutBanTime, "")

		msg := p.GetLocalString(lang, "on_dead", map[string]string{"USERNAME": username})
		p.SendMessage(channel, msg)

		p.startRecoverTimer(channel, username)
	})
}

func (p *Pinhead) startBoxTimer(channel string) {
	p.StopTimer(channel, BoxTimerName)

	chanState := p.GetState(channel)
	doctorSettings := chanState.Settings.Killers.Doctor
	lang := chanState.Settings.Language

	p.StartTimer(channel, BoxTimerName, doctorSettings.Timeout, func() {
		viewerList := p.GetViewerList(channel)

		p.UpdateState(channel, func(chanState *db.ChannelState) {
			for _, viewer := range viewerList {
				chanState.Stats["hits"]++
				chanState.UserMap[viewer].Health = "deep_wound"
				chanState.UserMap[viewer].Stats["hits"]++

				p.startDeadTimer(channel, viewer)
			}

			chanState.Killer = ""
			chanState.KillerState = nil
			chanState.Date = time.Now()
			chanState.Stats["success"]++
		})

		msg := p.GetLocalString(lang, "pinhead_success", map[string]string{})
		p.SendMessage(channel, msg)
	})
}

func (p *Pinhead) startBox(channel string) {
	startState := p.GetState(channel)
	lang := startState.Settings.Language
	pinheadSettings := startState.Settings.Killers.Pinhead
	now := time.Now()

	if startState.Killer != "" {
		slog.Warn("Killer is already summoned",
			slog.String("channel", channel),
		)
		return
	}

	genRes, err := p.GenerateWord(channel)
	if err != nil {
		slog.Error("Failed to generate a word for pinhead",
			slog.String("channel", channel),
			slog.Any("error", err),
		)
		return
	}

	p.UpdateState(channel, func(channelState *db.ChannelState) {
		channelState.Killer = "pinhead"
		channelState.KillerState = db.PinheadState{
			Word: genRes.Word,
		}
		channelState.Date = now
		channelState.Stats["total"]++
	})

	if pinheadSettings.ShowTopic {
		msg := p.GetLocalString(lang, "start_pinhead", map[string]string{"TOPIC": genRes.Topic})
		p.SendMessage(channel, msg)
	} else {
		msg := p.GetLocalString(lang, "start_pinhead_secret", nil)
		p.SendMessage(channel, msg)
	}

	p.startBoxTimer(channel)

	slog.Info("Box started",
		slog.String("channel", channel),
		slog.String("word", genRes.Word),
		slog.String("topic", genRes.Topic),
	)
}

func (p *Pinhead) Start(userMsg db.Message) {
	p.startBox(userMsg.Channel)
}

func (p *Pinhead) HandleMessage(userMsg db.Message) {
	chanState := p.GetState(userMsg.Channel)

	if chanState.Settings.Disabled {
		return
	}

	if p.handleCommands(userMsg) {
		return
	}
}

func (p *Pinhead) handleCommands(userMsg db.Message) bool {
	chanState := p.GetState(userMsg.Channel)
	lang := chanState.Settings.Language

	switch {
	case strings.HasPrefix(userMsg.Text, "!killer"):
		msg := p.GetLocalString(lang, "commands_pinhead", map[string]string{"STATS": fmt.Sprintf("https://leg.rofleksey.ru/#/stats/%s", userMsg.Channel)})
		p.SendMessage(userMsg.Channel, msg)
		return true
	case strings.HasPrefix(userMsg.Text, "!solve"):
		question := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(strings.ReplaceAll(userMsg.Text, "@", ""), "!solve")))

		var pinheadState db.PinheadState
		if err := mapstructure.Decode(chanState.KillerState, &pinheadState); err != nil {
			slog.Error("Failed to decode killer state",
				slog.String("channel", userMsg.Channel),
				slog.Any("error", err),
			)
		}

		res, err := p.GuessWord(lang, pinheadState.Word, question)
		if err != nil {
			slog.Error("Failed to guess word",
				slog.String("channel", userMsg.Channel),
				slog.String("word", pinheadState.Word),
				slog.String("question", question),
				slog.Any("error", err),
			)
		}
		if err != nil {
			slog.Error("Failed to guess word",
				slog.String("channel", userMsg.Channel),
				slog.String("word", pinheadState.Word),
				slog.String("question", question),
				slog.Any("error", err),
			)
		}

		switch res {
		case GuessResultOK:
			p.UpdateState(userMsg.Channel, func(chanState *db.ChannelState) {
				chanState.Killer = ""
				chanState.KillerState = nil
				chanState.Date = time.Now()
				chanState.Stats["fail"]++
			})

			msg := p.GetLocalString(lang, "pinhead_failure", map[string]string{"USERNAME": userMsg.Username, "WORD": pinheadState.Word})
			p.SendMessage(userMsg.Channel, msg)

			return true
		case GuessResultYes:
			msg := p.GetLocalString(lang, "pinhead_yes", map[string]string{"QUESTION": question, "USERNAME": userMsg.Username})
			p.SendMessage(userMsg.Channel, msg)

			return true
		case GuessResultNo:
			msg := p.GetLocalString(lang, "pinhead_no", map[string]string{"QUESTION": question, "USERNAME": userMsg.Username})
			p.SendMessage(userMsg.Channel, msg)

			return true
		case GuessResultMaybe:
			msg := p.GetLocalString(lang, "pinhead_maybe", map[string]string{"QUESTION": question, "USERNAME": userMsg.Username})
			p.SendMessage(userMsg.Channel, msg)

			return true
		case GuessResultIncorrect:
			msg := p.GetLocalString(lang, "pinhead_incorrect", map[string]string{"QUESTION": question, "USERNAME": userMsg.Username})
			p.SendMessage(userMsg.Channel, msg)

			return true
		}

		return true
	}

	return false
}
