package ghostface

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/samber/do"
	"legion-bot-v2/bot/i18n"
	"legion-bot-v2/bot/killer"
	"legion-bot-v2/chat"
	"legion-bot-v2/db"
	"legion-bot-v2/gpt"
	"legion-bot-v2/util"
	"legion-bot-v2/util/timers"
	"log/slog"
	"math/rand/v2"
	"strings"
	"time"
)

var _ killer.Killer = (*GhostFace)(nil)

const (
	StalkTimerName = "!!gf_stalk!!"
)

type GhostFace struct {
	db.DB
	chat.Actions
	timers.Timers
	i18n.Localiser
	gpt.Gpt
}

func New(di *do.Injector) *GhostFace {
	return &GhostFace{
		DB:        do.MustInvoke[db.DB](di),
		Actions:   do.MustInvoke[chat.Actions](di),
		Timers:    do.MustInvoke[timers.Timers](di),
		Localiser: do.MustInvoke[i18n.Localiser](di),
		Gpt:       do.MustInvoke[gpt.Gpt](di),
	}
}

func (g *GhostFace) HandleWhisper(userMsg db.PartialMessage) {

}

func (g *GhostFace) Name() string {
	return "ghostface"
}

func (g *GhostFace) Weight(channel string) int {
	chanState := g.GetState(channel)
	return chanState.Settings.Killers.GhostFace.Weight
}

func (g *GhostFace) Enabled(channel string) bool {
	chanState := g.GetState(channel)
	return chanState.Settings.Killers.GhostFace.Enabled
}

func (g *GhostFace) FixSettings(chanState *db.ChannelState) bool {
	if chanState.Settings.Killers.GhostFace != nil {
		return false
	}

	chanState.Settings.Killers.GhostFace = db.DefaultGhostFaceSettings()

	return true
}

func (g *GhostFace) startStalkTimer(channel string) {
	g.StopTimer(channel, StalkTimerName)

	chanState := g.GetState(channel)
	gfSettings := chanState.Settings.Killers.GhostFace
	lang := chanState.Settings.Language

	g.StartTimer(channel, StalkTimerName, gfSettings.Timeout, func() {
		chanState := g.GetState(channel)

		var gfState db.GhostFaceState
		if err := mapstructure.Decode(chanState.KillerState, &gfState); err != nil {
			slog.Error("Failed to decode killer state",
				slog.String("channel", channel),
				slog.Any("error", err),
			)
		}

		g.UpdateState(channel, func(chanState *db.ChannelState) {
			chanState.Killer = ""
			chanState.KillerState = nil
			chanState.Date = time.Now()
			chanState.Stats["fail"]++
		})

		msg := g.GetLocalString(lang, "gf_go_away", map[string]string{"COUNT": fmt.Sprint(len(gfState.StalkedThisRound))})
		g.SendMessage(channel, msg)
	})
}

func (g *GhostFace) startStalk(channel string) {
	startState := g.GetState(channel)
	lang := startState.Settings.Language
	now := time.Now()

	if startState.Killer != "" {
		slog.Warn("Killer is already summoned",
			slog.String("channel", channel),
		)
		return
	}

	g.UpdateState(channel, func(channelState *db.ChannelState) {
		channelState.Killer = "ghostface"
		channelState.Date = now
		channelState.Stats["total"]++
		channelState.KillerState = db.GhostFaceState{
			StalkedThisRound: make(map[string]bool),
		}
	})

	msg := g.GetLocalString(lang, "start_gf", nil)
	g.SendMessage(channel, msg)

	g.startStalkTimer(channel)

	slog.Info("Stalk started (ghostface)", slog.String("channel", channel))
}

func (g *GhostFace) Start(userMsg db.Message) {
	g.startStalk(userMsg.Channel)
}

func (g *GhostFace) HandleMessage(userMsg db.Message) {
	chanState := g.GetState(userMsg.Channel)
	gfSettings := chanState.Settings.Killers.GhostFace
	now := time.Now()

	if chanState.Settings.Disabled {
		return
	}

	if g.handleCommands(userMsg) {
		return
	}

	user := chanState.UserMap[userMsg.Username]
	diff := now.Sub(chanState.Date)

	if diff < gfSettings.MinDelayBetweenHits {
		return
	}

	if rand.Float64() > gfSettings.ReactChance {
		return
	}

	if user.Health == "hooked" || user.Health == "dead" {
		return
	}

	if userMsg.Username == userMsg.Channel || userMsg.IsMod || strings.Contains(userMsg.Username, "bot") || userMsg.Username == util.BotOwner {
		return
	}

	g.handleHit(userMsg.Channel, userMsg.Username)
}

func (g *GhostFace) TimeRemaining(channel string) time.Duration {
	return g.GetRemainingTime(channel, StalkTimerName)
}

func (g *GhostFace) handleCommands(userMsg db.Message) bool {
	chanState := g.GetState(userMsg.Channel)
	gfSettings := chanState.Settings.Killers.GhostFace
	lang := chanState.Settings.Language
	user := chanState.UserMap[userMsg.Username]

	switch {
	case strings.HasPrefix(userMsg.Text, "!killer"):
		msg := g.GetLocalString(lang, "commands_gf", map[string]string{"STATS": fmt.Sprintf("https://leg.rofleksey.ru/#/stats/%s", userMsg.Channel)})
		g.SendMessage(userMsg.Channel, msg)
		return true

	case strings.HasPrefix(userMsg.Text, "!tbag"):
		if user.Health == "hooked" || user.Health == "dead" {
			msg := g.GetLocalString(lang, "cant_do_rn", map[string]string{"USERNAME": userMsg.Username})
			g.SendMessage(userMsg.Channel, msg)
			return true
		}

		msg := g.GetLocalString(lang, "gf_tbag", map[string]string{"USERNAME": userMsg.Username})
		g.SendMessage(userMsg.Channel, msg)

		g.handleHit(userMsg.Channel, userMsg.Username)
		return true

	case strings.HasPrefix(userMsg.Text, "!reveal"):
		if user.Health == "hooked" || user.Health == "dead" || user.Marked {
			msg := g.GetLocalString(lang, "cant_do_rn", map[string]string{"USERNAME": userMsg.Username})
			g.SendMessage(userMsg.Channel, msg)
			return true
		}

		if rand.Float64() > gfSettings.RevealChance {
			msg := g.GetLocalString(lang, "gf_reveal_fail", map[string]string{"USERNAME": userMsg.Username})
			g.SendMessage(userMsg.Channel, msg)

			g.handleHit(userMsg.Channel, userMsg.Username)
			return true
		}

		msg := g.GetLocalString(lang, "gf_reveal", map[string]string{"USERNAME": userMsg.Username})
		g.SendMessage(userMsg.Channel, msg)

		g.handleHit(userMsg.Channel, userMsg.Username)

		chanState := g.GetState(userMsg.Channel)
		if chanState.Killer != "ghostface" {
			return true
		}

		var gfState db.GhostFaceState
		if err := mapstructure.Decode(chanState.KillerState, &gfState); err != nil {
			slog.Error("Failed to decode killer state",
				slog.String("channel", userMsg.Channel),
				slog.Any("error", err),
			)
		}

		g.UpdateState(userMsg.Channel, func(chanState *db.ChannelState) {
			chanState.Killer = ""
			chanState.KillerState = nil
			chanState.Date = time.Now()
			chanState.UserMap[userMsg.Username].Stats["stuns"]++
			chanState.Stats["fail"]++

			for u := range chanState.UserMap {
				if !gfState.StalkedThisRound[u] {
					chanState.UserMap[u].Marked = false
				}
			}
		})

		g.StopTimer(userMsg.Channel, StalkTimerName)

		msg = g.GetLocalString(lang, "gf_revealed", map[string]string{"USERNAME": userMsg.Username})
		g.SendMessage(userMsg.Channel, msg)

		msg = g.GetLocalString(lang, "gf_go_away", map[string]string{"COUNT": fmt.Sprint(len(gfState.StalkedThisRound))})
		g.SendMessage(userMsg.Channel, msg)

		return true
	}

	return false
}

func (g *GhostFace) handleHit(channel, username string) {
	chanState := g.GetState(channel)
	gfSettings := chanState.Settings.Killers.GhostFace
	lang := chanState.Settings.Language
	now := time.Now()

	var gfState db.GhostFaceState
	if err := mapstructure.Decode(chanState.KillerState, &gfState); err != nil {
		slog.Error("Failed to decode killer state",
			slog.String("channel", channel),
			slog.Any("error", err),
		)
	}

	user, userExists := chanState.UserMap[username]
	if !userExists {
		user = db.NewUser()
		g.UpdateState(channel, func(chanState *db.ChannelState) {
			chanState.UserMap[username] = user
			chanState.Date = now
		})
	}

	if gfState.StalkedThisRound[username] {
		return
	}

	if !user.Marked {
		gfState.StalkedThisRound[username] = true
		g.UpdateState(channel, func(chanState *db.ChannelState) {
			chanState.KillerState = gfState
			chanState.Date = now
			chanState.UserMap[username].Marked = true
		})
		return
	}

	g.UpdateState(channel, func(chanState *db.ChannelState) {
		chanState.Killer = ""
		chanState.KillerState = nil
		chanState.Date = now
		chanState.UserMap[username].Health = "hooked"
		chanState.UserMap[username].Marked = false
		chanState.UserMap[username].Stats["hooks"]++
		chanState.Stats["success"]++

		for u := range chanState.UserMap {
			if !gfState.StalkedThisRound[u] {
				chanState.UserMap[u].Marked = false
			}
		}
	})

	g.StopTimer(channel, StalkTimerName)
	g.StopTimer(channel, username)
	g.TimeoutUser(channel, username, gfSettings.HookBanTime, "")

	msg := g.GetLocalString(lang, "gf_hit_dead", map[string]string{"USERNAME": username})
	g.SendMessage(channel, msg)

	msg = g.GetLocalString(lang, "gf_go_away", map[string]string{"COUNT": fmt.Sprint(len(gfState.StalkedThisRound))})
	g.SendMessage(channel, msg)
}
