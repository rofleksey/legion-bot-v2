package bot

import (
	"github.com/elliotchance/pie/v2"
	"github.com/jellydator/ttlcache/v3"
	"legion-bot-v2/chat"
	"legion-bot-v2/db"
	"legion-bot-v2/i18n"
	"legion-bot-v2/killer"
	"legion-bot-v2/timers"
	"log/slog"
	"math/rand"
	"strings"
	"time"
)

type Bot struct {
	db.DB
	chat.Actions
	timers.Timers
	i18n.Localiser
	killerMap      map[string]killer.Killer
	streamStartMap *ttlcache.Cache[string, time.Time]
	viewerCountMap *ttlcache.Cache[string, int]
}

func NewBot(
	db db.DB,
	actions chat.Actions,
	timers timers.Timers,
	localiser i18n.Localiser,
	killerMap map[string]killer.Killer,
) *Bot {
	bot := &Bot{
		DB:        db,
		Actions:   actions,
		Timers:    timers,
		Localiser: localiser,
		killerMap: killerMap,
		streamStartMap: ttlcache.New[string, time.Time](
			ttlcache.WithTTL[string, time.Time](30 * time.Minute),
		),
		viewerCountMap: ttlcache.New[string, int](
			ttlcache.WithTTL[string, int](5*time.Minute),
			ttlcache.WithDisableTouchOnHit[string, int](),
		),
	}

	return bot
}

func (b *Bot) Init() {
	for _, chanState := range b.GetAllStates() {
		if chanState.Killer != "" {
			b.UpdateState(chanState.Channel, func(chanState *db.ChannelState) {
				chanState.Killer = ""
				chanState.KillerState = nil
				chanState.Date = time.Now()
			})
		}

		if chanState.Settings.Killers.General == nil {
			b.UpdateState(chanState.Channel, func(chanState *db.ChannelState) {
				chanState.Settings.Killers.General = db.DefaultGeneralKillerSettings()
			})
		}

		for _, k := range b.killerMap {
			k.FixSettings(chanState.Channel)
		}

		for username, user := range chanState.UserMap {
			if user.Health == "dead" || user.Health == "deep_wound" {
				b.UpdateState(chanState.Channel, func(chanState *db.ChannelState) {
					chanState.UserMap[username].Health = "injured"
				})
			}
		}
	}
}

func (b *Bot) HandleCommands(userMsg db.Message) bool {
	chanState := b.GetState(userMsg.Channel)
	lang := chanState.Settings.Language
	user := chanState.UserMap[userMsg.Username]

	switch {
	case strings.HasPrefix(userMsg.Text, "!hp"):
		otherUsername := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(strings.ReplaceAll(userMsg.Text, "@", ""), "!hp")))
		if otherUsername == "" {
			otherUsername = userMsg.Username
		}

		otherUser, userExists := chanState.UserMap[otherUsername]
		if !userExists {
			otherUser = db.NewUser()
			b.UpdateState(userMsg.Channel, func(chanState *db.ChannelState) {
				chanState.UserMap[otherUsername] = otherUser
			})
		}

		var msg string
		switch otherUser.Health {
		case "hooked":
			msg = b.GetLocalString(lang, "hooked", map[string]string{"USERNAME": otherUsername})
		case "deep_wound":
			msg = b.GetLocalString(lang, "deep_wound", map[string]string{"USERNAME": otherUsername})
		case "injured":
			msg = b.GetLocalString(lang, "injured", map[string]string{"USERNAME": otherUsername})
		case "dead":
			msg = b.GetLocalString(lang, "dead", map[string]string{"USERNAME": otherUsername})
		case "healthy":
			msg = b.GetLocalString(lang, "healthy", map[string]string{"USERNAME": otherUsername})
		default:
			return true
		}

		b.SendMessage(userMsg.Channel, msg)

		return true

	case strings.HasPrefix(userMsg.Text, "!unhook"):
		otherUsername := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(strings.ReplaceAll(userMsg.Text, "@", ""), "!unhook")))
		if otherUsername == "" {
			otherUsername = userMsg.Username
		}

		otherUser, userExists := chanState.UserMap[otherUsername]
		if !userExists {
			otherUser = db.NewUser()
			b.UpdateState(userMsg.Channel, func(chanState *db.ChannelState) {
				chanState.UserMap[otherUsername] = otherUser
			})
		}

		if otherUsername == userMsg.Username {
			msg := b.GetLocalString(lang, "cant_unhook_self", map[string]string{"USERNAME": otherUsername})
			b.SendMessage(userMsg.Channel, msg)

			return true
		}

		if otherUser.Health != "hooked" {
			msg := b.GetLocalString(lang, "not_hooked", map[string]string{"USERNAME": otherUsername})
			b.SendMessage(userMsg.Channel, msg)

			return true
		}

		b.UpdateState(userMsg.Channel, func(chanState *db.ChannelState) {
			chanState.UserMap[otherUsername].Health = "healthy"

			chanState.UserMap[userMsg.Username].Stats["unhooks"]++
		})

		b.UnbanUser(userMsg.Channel, otherUsername)
		b.StopTimer(userMsg.Channel, otherUsername)

		msg := b.GetLocalString(lang, "on_unhooked", map[string]string{"USERNAME": otherUsername})
		b.SendMessage(userMsg.Channel, msg)

		return true

	case strings.HasPrefix(userMsg.Text, "!heal"):
		otherUsername := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(strings.ReplaceAll(userMsg.Text, "@", ""), "!heal")))
		if otherUsername == "" {
			otherUsername = userMsg.Username
		}

		otherUser, userExists := chanState.UserMap[otherUsername]
		if !userExists {
			otherUser = db.NewUser()
			b.UpdateState(userMsg.Channel, func(chanState *db.ChannelState) {
				chanState.UserMap[otherUsername] = otherUser
			})
		}

		if otherUsername == userMsg.Username {
			msg := b.GetLocalString(lang, "cant_heal_self", map[string]string{"USERNAME": otherUsername})
			b.SendMessage(userMsg.Channel, msg)
			return true
		}

		if user.Health == "hooked" || user.Health == "dead" {
			msg := b.GetLocalString(lang, "cant_do_rn", map[string]string{"USERNAME": otherUsername})
			b.SendMessage(userMsg.Channel, msg)
			return true
		}

		if otherUser.Health == "hooked" {
			msg := b.GetLocalString(lang, "hooked", map[string]string{"USERNAME": otherUsername})
			b.SendMessage(userMsg.Channel, msg)
			return true
		}

		if otherUser.Health == "healthy" {
			msg := b.GetLocalString(lang, "healthy", map[string]string{"USERNAME": otherUsername})
			b.SendMessage(userMsg.Channel, msg)
			return true
		}

		if otherUser.Health == "dead" {
			b.UnbanUser(userMsg.Channel, otherUsername)
		}

		b.UpdateState(userMsg.Channel, func(chanState *db.ChannelState) {
			chanState.UserMap[otherUsername].Health = "healthy"
			chanState.UserMap[userMsg.Username].Stats["heals"]++
		})

		b.StopTimer(userMsg.Channel, otherUsername)

		msg := b.GetLocalString(lang, "on_heal", map[string]string{"USERNAME": otherUsername})
		b.SendMessage(userMsg.Channel, msg)

		return true

	case strings.HasPrefix(userMsg.Text, "!mend"):
		if user.Health != "deep_wound" {
			msg := b.GetLocalString(lang, "not_deep_wound", map[string]string{"USERNAME": userMsg.Username})
			b.SendMessage(userMsg.Channel, msg)

			return true
		}

		b.UpdateState(userMsg.Channel, func(chanState *db.ChannelState) {
			chanState.UserMap[userMsg.Username].Health = "injured"
		})

		b.StopTimer(userMsg.Channel, userMsg.Username)

		msg := b.GetLocalString(lang, "on_mend", map[string]string{"USERNAME": userMsg.Username})
		b.SendMessage(userMsg.Channel, msg)

		return true
	}

	return false
}

func (b *Bot) getCachedStreamStartTime(channel string) time.Time {
	item := b.streamStartMap.Get(channel)
	if item != nil {
		return item.Value()
	}

	startTime := b.GetStartTime(channel)
	if !startTime.IsZero() {
		b.streamStartMap.Set(channel, startTime, ttlcache.DefaultTTL)
	} else {
		b.streamStartMap.Set(channel, startTime, 5*time.Minute)
	}

	return startTime
}

func (b *Bot) getCachedViewerCount(channel string) int {
	item := b.viewerCountMap.Get(channel)
	if item != nil {
		return item.Value()
	}

	count := b.GetViewerCount(channel)
	b.viewerCountMap.Set(channel, count, ttlcache.DefaultTTL)

	return count
}

func (b *Bot) HandleMessage(userMsg db.Message) {
	chanState := b.GetState(userMsg.Channel)
	generalKillerSettings := chanState.Settings.Killers.General

	if chanState.Settings.Disabled {
		return
	}

	streamStartTime := b.getCachedStreamStartTime(userMsg.Channel)

	var streamLength time.Duration
	if !streamStartTime.IsZero() {
		streamLength = time.Now().Sub(streamStartTime)
	}

	user, userExists := chanState.UserMap[userMsg.Username]
	if !userExists {
		user = db.NewUser()
		b.UpdateState(userMsg.Channel, func(chanState *db.ChannelState) {
			chanState.UserMap[userMsg.Username] = user
		})
	}

	if b.HandleCommands(userMsg) {
		return
	}

	diff := time.Now().Sub(chanState.Date)

	if chanState.Killer == "" {
		if diff <= generalKillerSettings.DelayBetweenKillers || streamLength <= generalKillerSettings.DelayAtTheStreamStart {
			return
		}

		killerList := pie.Filter(pie.Values(b.killerMap), func(k killer.Killer) bool {
			return k.Enabled(userMsg.Channel)
		})

		if len(killerList) == 0 {
			return
		}

		viewerCount := b.getCachedViewerCount(userMsg.Channel)
		if viewerCount < generalKillerSettings.MinNumberOfViewers {
			return
		}

		nextKiller := selectKiller(killerList, userMsg.Channel)
		nextKiller.Start(userMsg)
		return
	}

	curKiller, ok := b.killerMap[chanState.Killer]
	if !ok {
		slog.Error("Killer not found",
			slog.String("channel", chanState.Channel),
			slog.String("killer", chanState.Killer),
		)
		return
	}

	curKiller.HandleMessage(userMsg)
}

func selectKiller(arr []killer.Killer, channel string) killer.Killer {
	totalWeight := 0
	for _, k := range arr {
		totalWeight += k.Weight(channel)
	}

	r := rand.Intn(totalWeight)

	runningTotal := 0
	for _, k := range arr {
		runningTotal += k.Weight(channel)
		if r < runningTotal {
			return k
		}
	}

	return arr[0]
}
