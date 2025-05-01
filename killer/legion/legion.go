package legion

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"legion-bot-v2/chat"
	"legion-bot-v2/db"
	"legion-bot-v2/gpt"
	"legion-bot-v2/i18n"
	"legion-bot-v2/killer"
	"legion-bot-v2/timers"
	"legion-bot-v2/util"
	"log/slog"
	"math/rand/v2"
	"strings"
	"time"
)

var _ killer.Killer = (*Legion)(nil)

const (
	FrenzyTimerName = "!!frenzy!!"
)

type Legion struct {
	db.DB
	chat.Actions
	timers.Timers
	i18n.Localiser
	gpt.Gpt
}

func (l *Legion) HandleWhisper(userMsg db.PartialMessage) {

}

func (l *Legion) Weight(channel string) int {
	chanState := l.GetState(channel)
	return chanState.Settings.Killers.Legion.Weight
}

func (l *Legion) FixSettings(chanState *db.ChannelState) bool {
	if chanState.Settings.Killers.Legion != nil {
		return false
	}

	chanState.Settings.Killers.Legion = db.DefaultLegionSettings()

	return true
}

func (l *Legion) Name() string {
	return "legion"
}

func (l *Legion) Enabled(channel string) bool {
	chanState := l.GetState(channel)
	return chanState.Settings.Killers.Legion.Enabled
}

func New(db db.DB, actions chat.Actions, timers timers.Timers, localiser i18n.Localiser, g gpt.Gpt) *Legion {
	k := &Legion{
		DB:        db,
		Actions:   actions,
		Timers:    timers,
		Localiser: localiser,
		Gpt:       g,
	}

	return k
}

func (l *Legion) handleCommands(userMsg db.Message) bool {
	chanState := l.GetState(userMsg.Channel)
	legionSettings := chanState.Settings.Killers.Legion
	lang := chanState.Settings.Language
	now := time.Now()
	user := chanState.UserMap[userMsg.Username]

	switch {
	case strings.HasPrefix(userMsg.Text, "!killer"):
		msg := l.GetLocalString(lang, "commands_legion", map[string]string{"STATS": fmt.Sprintf("https://leg.rofleksey.ru/#/stats/%s", userMsg.Channel)})
		l.SendMessage(userMsg.Channel, msg)
		return true

	case strings.HasPrefix(userMsg.Text, "!pallet"):
		if user.Health == "hooked" || user.Health == "dead" {
			msg := l.GetLocalString(lang, "cant_do_rn", map[string]string{"USERNAME": userMsg.Username})
			l.SendMessage(userMsg.Channel, msg)

			return true
		}

		if chanState.Killer == "" || user.Health == "deep_wound" {
			msg := l.GetLocalString(lang, "pallet_wasted", map[string]string{"USERNAME": userMsg.Username})
			l.SendMessage(userMsg.Channel, msg)

			return true
		}

		if rand.Float64() > legionSettings.PalletStunChance {
			msg := l.GetLocalString(lang, "pallet_failed", map[string]string{"USERNAME": userMsg.Username})
			l.SendMessage(userMsg.Channel, msg)

			l.handleHit(userMsg.Channel, userMsg.Username)
			return true
		}

		l.UpdateState(userMsg.Channel, func(chanState *db.ChannelState) {
			chanState.Killer = ""
			chanState.KillerState = nil
			chanState.Date = now
			chanState.Stats["stuns"]++
			chanState.UserMap[userMsg.Username].Stats["stuns"]++
		})

		l.StopTimer(userMsg.Channel, FrenzyTimerName)

		msg := l.GetLocalString(lang, "pallet_success", map[string]string{"USERNAME": userMsg.Username})
		l.SendMessage(userMsg.Channel, msg)
		return true

	case strings.HasPrefix(userMsg.Text, "!tbag"):
		if user.Health == "hooked" || user.Health == "dead" {
			msg := l.GetLocalString(lang, "cant_do_rn", map[string]string{"USERNAME": userMsg.Username})
			l.SendMessage(userMsg.Channel, msg)
			return true
		}

		if chanState.Killer == "" || user.Health == "deep_wound" {
			msg := l.GetLocalString(lang, "tbag_wasted", map[string]string{"USERNAME": userMsg.Username})
			l.SendMessage(userMsg.Channel, msg)

			return true
		}

		msg := l.GetLocalString(lang, "tbag_success", map[string]string{"USERNAME": userMsg.Username})
		l.SendMessage(userMsg.Channel, msg)

		l.handleHit(userMsg.Channel, userMsg.Username)
		return true

	case strings.HasPrefix(userMsg.Text, "!locker"):
		if user.Health == "hooked" || user.Health == "dead" {
			msg := l.GetLocalString(lang, "cant_do_rn", map[string]string{"USERNAME": userMsg.Username})
			l.SendMessage(userMsg.Channel, msg)

			return true
		}

		if chanState.Killer == "" || user.Health == "deep_wound" {
			msg := l.GetLocalString(lang, "locker_wasted", map[string]string{"USERNAME": userMsg.Username})
			l.SendMessage(userMsg.Channel, msg)
			return true
		}

		if rand.Float64() > legionSettings.LockerStunChance {
			msg := l.GetLocalString(lang, "locker_failed", map[string]string{"USERNAME": userMsg.Username})
			l.SendMessage(userMsg.Channel, msg)

			if rand.Float64() > legionSettings.LockerGrabChance {
				l.handleHit(userMsg.Channel, userMsg.Username)
				return true
			}

			l.UpdateState(userMsg.Channel, func(chanState *db.ChannelState) {
				chanState.Killer = ""
				chanState.KillerState = nil
				chanState.Date = now
				chanState.UserMap[userMsg.Username].Health = "hooked"
				chanState.UserMap[userMsg.Username].Stats["hooks"]++
				chanState.Stats["success"]++
			})

			l.StopTimer(userMsg.Channel, FrenzyTimerName)
			l.StopTimer(userMsg.Channel, userMsg.Username)
			l.TimeoutUser(userMsg.Channel, userMsg.Username, legionSettings.HookBanTime, "")

			msg = l.GetLocalString(lang, "locker_grab", map[string]string{"USERNAME": userMsg.Username})
			l.SendMessage(userMsg.Channel, msg)

			return true
		}

		l.UpdateState(userMsg.Channel, func(chanState *db.ChannelState) {
			chanState.Killer = ""
			chanState.KillerState = nil
			chanState.Date = now
			chanState.Stats["stuns"]++
			chanState.UserMap[userMsg.Username].Stats["stuns"]++
		})

		l.StopTimer(userMsg.Channel, FrenzyTimerName)

		msg := l.GetLocalString(lang, "locker_success", map[string]string{"USERNAME": userMsg.Username})
		l.SendMessage(userMsg.Channel, msg)

		return true
	}

	return false
}

func (l *Legion) Start(userMsg db.Message) {
	l.startFrenzy(userMsg.Channel)
}

func (l *Legion) HandleMessage(userMsg db.Message) {
	chanState := l.GetState(userMsg.Channel)
	legionSettings := chanState.Settings.Killers.Legion
	lang := chanState.Settings.Language
	now := time.Now()

	if chanState.Settings.Disabled {
		return
	}

	if l.handleCommands(userMsg) {
		return
	}

	user := chanState.UserMap[userMsg.Username]
	diff := now.Sub(chanState.Date)

	if diff < legionSettings.MinDelayBetweenHits {
		return
	}

	if rand.Float64() > legionSettings.ReactChance {
		return
	}

	if user.Health == "hooked" {
		msg := l.GetLocalString(lang, "on_hook_camp", map[string]string{"USERNAME": userMsg.Username})
		l.SendMessage(userMsg.Channel, msg)
		return
	}

	if user.Health == "dead" {
		msg := l.GetLocalString(lang, "on_dead_camp", map[string]string{"USERNAME": userMsg.Username})
		l.SendMessage(userMsg.Channel, msg)
		return
	}

	if userMsg.Username == userMsg.Channel || userMsg.IsMod || strings.Contains(userMsg.Username, "bot") || userMsg.Username == util.BotOwner {
		return
	}

	l.handleHit(userMsg.Channel, userMsg.Username)
}

func (l *Legion) startFrenzyTimer(channel string) {
	l.StopTimer(channel, FrenzyTimerName)

	chanState := l.GetState(channel)
	legionSettings := chanState.Settings.Killers.Legion
	lang := chanState.Settings.Language

	l.StartTimer(channel, FrenzyTimerName, legionSettings.FrenzyTimeout, func() {
		l.UpdateState(channel, func(chanState *db.ChannelState) {
			chanState.Killer = ""
			chanState.KillerState = nil
			chanState.Date = time.Now()
			chanState.Stats["fail"]++
		})

		msg := l.GetLocalString(lang, "frenzy_timeout", nil)
		l.SendMessage(channel, msg)
	})
}

func (l *Legion) startFrenzy(channel string) {
	startState := l.GetState(channel)
	lang := startState.Settings.Language
	now := time.Now()

	if startState.Killer != "" {
		slog.Warn("Killer is already summoned",
			slog.String("channel", channel),
		)
		return
	}

	l.UpdateState(channel, func(channelState *db.ChannelState) {
		channelState.Killer = "legion"
		channelState.Date = now
		channelState.Stats["total"]++
		channelState.KillerState = db.LegionState{
			HitCount: 0,
		}
	})

	msg := l.GetLocalString(lang, "start_legion", nil)
	l.SendMessage(channel, msg)

	l.startFrenzyTimer(channel)

	slog.Info("Frenzy started", slog.String("channel", channel))
}

func (l *Legion) startRecoverTimer(channel, username string) {
	l.StopTimer(channel, username)

	chanState := l.GetState(channel)
	legionSettings := chanState.Settings.Killers.Legion

	l.StartTimer(channel, username, legionSettings.BleedOutBanTime, func() {
		l.UpdateState(channel, func(chanState *db.ChannelState) {
			chanState.UserMap[username].Health = "injured"
		})
	})
}

func (l *Legion) startDeadTimer(channel, username string) {
	l.StopTimer(channel, username)

	chanState := l.GetState(channel)
	legionSettings := chanState.Settings.Killers.Legion
	lang := chanState.Settings.Language

	l.StartTimer(channel, username, legionSettings.DeepWoundTimeout, func() {
		l.UpdateState(channel, func(chanState *db.ChannelState) {
			chanState.Stats["bleedOuts"]++

			chanState.UserMap[username].Health = "dead"
			chanState.Stats["bleedOuts"]++
		})

		l.TimeoutUser(channel, username, legionSettings.BleedOutBanTime, "")

		msg := l.GetLocalString(lang, "on_dead", map[string]string{"USERNAME": username})
		l.SendMessage(channel, msg)

		l.startRecoverTimer(channel, username)
	})
}

func (l *Legion) handleHit(channel, username string) {
	chanState := l.GetState(channel)
	legionSettings := chanState.Settings.Killers.Legion
	lang := chanState.Settings.Language
	now := time.Now()

	var legionState db.LegionState
	if err := mapstructure.Decode(chanState.KillerState, &legionState); err != nil {
		slog.Error("Failed to decode killer state",
			slog.String("channel", channel),
			slog.Any("error", err),
		)
	}

	user, userExists := chanState.UserMap[username]
	if !userExists {
		user = db.NewUser()
		l.UpdateState(channel, func(chanState *db.ChannelState) {
			chanState.UserMap[username] = user
		})
	}

	if rand.Float64() > legionSettings.HitChance {
		l.UpdateState(channel, func(chanState *db.ChannelState) {
			chanState.Killer = ""
			chanState.KillerState = nil
			chanState.Date = now
			chanState.Stats["miss"]++
		})

		l.StopTimer(channel, FrenzyTimerName)

		msg := l.GetLocalString(lang, "on_frenzy_miss", map[string]string{"USERNAME": username})
		l.SendMessage(channel, msg)

		return
	}

	if legionState.HitCount == legionSettings.FatalHit-1 {
		l.UpdateState(channel, func(chanState *db.ChannelState) {
			chanState.Killer = ""
			chanState.KillerState = nil
			chanState.Date = now
			chanState.UserMap[username].Health = "hooked"
			chanState.UserMap[username].Stats["hooks"]++
			chanState.Stats["success"]++
		})

		l.StopTimer(channel, FrenzyTimerName)
		l.StopTimer(channel, username)
		l.TimeoutUser(channel, username, legionSettings.HookBanTime, "")

		msg := l.GetLocalString(lang, "on_frenzy_hit_dead", map[string]string{"USERNAME": username})
		l.SendMessage(channel, msg)

		return
	}

	if user.Health == "deep_wound" {
		if rand.Float64() > legionSettings.BodyBlockSuccessChance {
			return
		}

		l.UpdateState(channel, func(chanState *db.ChannelState) {
			chanState.Killer = ""
			chanState.KillerState = nil
			chanState.Date = now
			chanState.Stats["bodyBlock"]++
			chanState.UserMap[username].Stats["bodyBlocks"]++
		})

		l.StopTimer(channel, FrenzyTimerName)
		l.startDeadTimer(channel, username)

		msg := l.GetLocalString(lang, "on_frenzy_hit_deep_wound", map[string]string{"USERNAME": username})
		l.SendMessage(channel, msg)

		return
	}

	legionState.HitCount++
	l.UpdateState(channel, func(chanState *db.ChannelState) {
		chanState.KillerState = legionState
		chanState.Stats["hits"]++
		chanState.Date = now
		chanState.UserMap[username].Health = "deep_wound"
		chanState.UserMap[username].Stats["hits"]++
	})

	l.startDeadTimer(channel, username)
	l.startFrenzyTimer(channel)

	if legionState.HitCount == legionSettings.FatalHit-1 {
		msg := l.GetLocalString(lang, "on_frenzy_hit_prefinal", map[string]string{"USERNAME": username})
		l.SendMessage(channel, msg)
	} else {
		msg := l.GetLocalString(lang, "on_frenzy_hit", map[string]string{"USERNAME": username})
		l.SendMessage(channel, msg)
	}

	return
}
