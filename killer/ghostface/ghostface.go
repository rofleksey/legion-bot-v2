package ghostface

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
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
}

func New(db db.DB, actions chat.Actions, timers timers.Timers, localiser i18n.Localiser) *GhostFace {
	k := &GhostFace{
		DB:        db,
		Actions:   actions,
		Timers:    timers,
		Localiser: localiser,
	}

	return k
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

func (g *GhostFace) FixSettings(channel string) {
	chanState := g.GetState(channel)
	if chanState.Settings.Killers.GhostFace == nil {
		g.UpdateState(chanState.Channel, func(chanState *db.ChannelState) {
			chanState.Settings.Killers.GhostFace = db.DefaultGhostFaceSettings()
		})
	}
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
	lang := chanState.Settings.Language
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

	if user.Health == "hooked" {
		msg := g.GetLocalString(lang, "gf_on_hook_camp", map[string]string{"USERNAME": userMsg.Username})
		g.SendMessage(userMsg.Channel, msg)
		return
	}

	if user.Health == "dead" {
		return
	}

	if userMsg.Username == userMsg.Channel || userMsg.IsMod || strings.Contains(userMsg.Username, "bot") || userMsg.Username == util.BotOwner {
		return
	}

	g.handleHit(userMsg.Channel, userMsg.Username)
}

func (g *GhostFace) handleCommands(userMsg db.Message) bool {
	chanState := g.GetState(userMsg.Channel)
	lang := chanState.Settings.Language
	user := chanState.UserMap[userMsg.Username]

	switch {
	case strings.HasPrefix(userMsg.Text, "!killer"):
		msg := g.GetLocalString(lang, "commands_gf", map[string]string{"STATS": fmt.Sprintf("https://leg.rofleksey.ru/#/stats/%s", userMsg.Channel)})
		g.SendMessage(userMsg.Channel, msg)
		return true

	case strings.HasPrefix(userMsg.Text, "!tbag"):
		if user.Health == "hooked" || user.Health == "dead" {
			msg := g.GetLocalString(lang, "cant_tbag_rn", map[string]string{"USERNAME": userMsg.Username})
			g.SendMessage(userMsg.Channel, msg)
			return true
		}

		msg := g.GetLocalString(lang, "gf_tbag", map[string]string{"USERNAME": userMsg.Username})
		g.SendMessage(userMsg.Channel, msg)

		g.handleHit(userMsg.Channel, userMsg.Username)
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
		chanState.UserMap[username].Stats["hooks"]++
		chanState.Stats["success"]++
	})

	g.StopTimer(channel, StalkTimerName)
	g.StopTimer(channel, username)
	g.TimeoutUser(channel, username, gfSettings.HookBanTime, "")

	msg := g.GetLocalString(lang, "gf_hit_dead", map[string]string{"USERNAME": username})
	g.SendMessage(channel, msg)

	msg = g.GetLocalString(lang, "gf_go_away", map[string]string{"COUNT": fmt.Sprint(len(gfState.StalkedThisRound))})
	g.SendMessage(channel, msg)
}
