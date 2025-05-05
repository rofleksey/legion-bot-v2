package bot

import (
	_ "embed"
	"fmt"
	"github.com/elliotchance/pie/v2"
	"github.com/jellydator/ttlcache/v3"
	"github.com/samber/do"
	"legion-bot-v2/api/dao"
	"legion-bot-v2/bot/i18n"
	"legion-bot-v2/bot/killer"
	"legion-bot-v2/db"
	"legion-bot-v2/gpt"
	"legion-bot-v2/twitch/chat"
	"legion-bot-v2/util"
	"legion-bot-v2/util/timers"
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
	gpt.Gpt
	killerMap      map[string]killer.Killer
	streamStartMap *ttlcache.Cache[string, time.Time]
	viewerCountMap *ttlcache.Cache[string, int]
}

func NewBot(di *do.Injector) *Bot {
	streamStartMap := ttlcache.New[string, time.Time](
		ttlcache.WithTTL[string, time.Time](30 * time.Minute),
	)
	go streamStartMap.Start()

	viewerCountMap := ttlcache.New[string, int](
		ttlcache.WithTTL[string, int](5*time.Minute),
		ttlcache.WithDisableTouchOnHit[string, int](),
	)
	go viewerCountMap.Start()

	bot := &Bot{
		DB:             do.MustInvoke[db.DB](di),
		Actions:        do.MustInvoke[chat.Actions](di),
		Timers:         do.MustInvoke[timers.Timers](di),
		Localiser:      do.MustInvoke[i18n.Localiser](di),
		Gpt:            do.MustInvoke[gpt.Gpt](di),
		killerMap:      do.MustInvoke[map[string]killer.Killer](di),
		streamStartMap: streamStartMap,
		viewerCountMap: viewerCountMap,
	}

	return bot
}

func (b *Bot) Init() {
	channels := b.GetAllChannelNames()

	for _, channel := range channels {
		b.UpdateState(channel, func(chanState *db.ChannelState) {
			if chanState.Killer != "" {
				chanState.Killer = ""
				chanState.KillerState = nil
				chanState.Date = time.Now()
			}

			if chanState.Settings.Killers.General == nil {
				chanState.Settings.Killers.General = db.DefaultGeneralKillerSettings()
			}

			for _, k := range b.killerMap {
				k.FixSettings(chanState)
			}

			for username, user := range chanState.UserMap {
				if user.Health == "dead" || user.Health == "deep_wound" {
					chanState.UserMap[username].Health = "injured"
				}
			}
		})
	}
}

func (b *Bot) GetCachedStreamStartTime(channel string) time.Time {
	item := b.streamStartMap.Get(channel)
	if item != nil {
		return item.Value()
	}

	startTime := b.GetStartTime(channel)
	if !startTime.IsZero() {
		b.streamStartMap.Set(channel, startTime, ttlcache.DefaultTTL)
	}

	return startTime
}

func (b *Bot) GetCachedViewerCount(channel string) int {
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

	if chanState.Settings.Disabled || time.Now().Before(chanState.UserTimeout) {
		slog.Debug("HandleMessage ignored",
			slog.String("channel", userMsg.Channel),
			slog.String("cause", "bot disabled"),
		)
		return
	}

	streamStartTime := b.GetCachedStreamStartTime(userMsg.Channel)

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
			slog.Debug("startRandomKiller ignored",
				slog.String("channel", userMsg.Channel),
				slog.String("cause", "ongoing delay"),
			)
			return
		}

		b.startRandomKiller(userMsg)
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

func (b *Bot) HandleStreamOnline(channel string) {
	chanState := b.GetState(channel)
	lang := chanState.Settings.Language

	if chanState.Settings.Disabled || time.Now().Before(chanState.UserTimeout) {
		slog.Debug("HandleStreamOnline ignored",
			slog.String("channel", channel),
			slog.String("cause", "bot is disabled"),
		)
		return
	}

	b.GetCachedStreamStartTime(channel)

	msg := b.GetLocalString(lang, "stream_start_greeting", map[string]string{})
	b.SendMessage(channel, msg)
}

func (b *Bot) HandleStreamOffline(channel string) {
	chanState := b.GetState(channel)
	lang := chanState.Settings.Language

	if chanState.Settings.Disabled || time.Now().Before(chanState.UserTimeout) {
		slog.Debug("HandleStreamOffline ignored",
			slog.String("channel", channel),
			slog.String("cause", "bot is disabled"),
		)
		return
	}

	b.streamStartMap.Delete(channel)

	msg := b.GetLocalString(lang, "stream_end_greeting", map[string]string{})
	b.SendMessage(channel, msg)
}

func (b *Bot) HandleWhisper(username, message string) {
	channels := b.GetAllChannelNames()

	for _, channel := range channels {
		chanState := b.GetState(channel)

		if chanState.Settings.Disabled ||
			time.Now().Before(chanState.UserTimeout) ||
			chanState.Killer == "" {
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

		curKiller.HandleWhisper(db.PartialMessage{
			Channel:  channel,
			Username: username,
			Text:     message,
		})
	}
}

func (b *Bot) HandleIncomingRaid(channel, otherChannel string) {
	chanState := b.GetState(channel)
	chatSettings := chanState.Settings.Chat

	if !chatSettings.StartKillerOnRaid || chanState.Killer != "" {
		return
	}

	b.startRandomKiller(db.Message{
		Channel:  chanState.Channel,
		Username: util.BotUsername,
		IsMod:    false,
		Text:     "",
	})
}

func (b *Bot) HandleOutgoingRaid(channel, otherChannel string) {
	chanState := b.GetState(channel)

	chatSettings := chanState.Settings.Chat
	followRaids := chatSettings.FollowRaids
	followRaidsMessage := strings.TrimSpace(chatSettings.FollowRaidsMessage)

	if chanState.Settings.Disabled ||
		time.Now().Before(chanState.UserTimeout) ||
		!followRaids ||
		followRaidsMessage == "" {
		return
	}

	b.SendForeignMessage(otherChannel, followRaidsMessage)
}

func (b *Bot) HandleNewSteamComment(channel string, comment dao.Comment) {
	chanState := b.GetState(channel)
	lang := chanState.Settings.Language
	steamSettings := chanState.Settings.Steam

	if chanState.Settings.Disabled ||
		time.Now().Before(chanState.UserTimeout) ||
		!steamSettings.NotifyNewComments {
		return
	}

	msg := b.GetLocalString(lang, "steam_new_comment", map[string]string{"CHANNEL": channel})
	b.SendMessage(channel, msg)
}

func (b *Bot) startRandomKiller(userMsg db.Message) {
	slog.Debug("Starting random killer",
		slog.String("channel", userMsg.Channel),
	)

	chanState := b.GetState(userMsg.Channel)
	generalKillerSettings := chanState.Settings.Killers.General

	if chanState.Killer != "" {
		slog.Debug("Failed to start random killer",
			slog.String("channel", userMsg.Channel),
			slog.String("cause", "killer is already running"),
		)
		return
	}

	killerList := pie.Filter(pie.Values(b.killerMap), func(k killer.Killer) bool {
		return k.Enabled(userMsg.Channel)
	})

	if len(killerList) == 0 {
		slog.Debug("Failed to start random killer",
			slog.String("channel", userMsg.Channel),
			slog.String("cause", "no killers are enabled"),
		)
		return
	}

	viewerCount := b.GetCachedViewerCount(userMsg.Channel)
	if viewerCount < generalKillerSettings.MinNumberOfViewers {
		slog.Debug("Failed to start random killer",
			slog.String("channel", userMsg.Channel),
			slog.String("cause", "viewer count is too small"),
			slog.Int("cur_count", viewerCount),
			slog.Int("required_count", generalKillerSettings.MinNumberOfViewers),
		)
		return
	}

	nextKiller := selectKillerWeighted(killerList, userMsg.Channel)

	slog.Debug("Starting killer",
		slog.String("channel", userMsg.Channel),
		slog.String("name", nextKiller.Name()),
	)

	nextKiller.Start(userMsg)
}

func (b *Bot) StartSpecificKiller(channel, name string) error {
	chanState := b.GetState(channel)

	if chanState.Killer != "" {
		return fmt.Errorf("killer is already running")
	}

	if chanState.Settings.Disabled || time.Now().Before(chanState.UserTimeout) {
		return fmt.Errorf("bot is disabled")
	}

	nextKiller, ok := b.killerMap[name]
	if !ok {
		slog.Error("Killer not found",
			slog.String("channel", channel),
			slog.String("name", name),
		)
		return fmt.Errorf("killer not found")
	}

	slog.Debug("Starting killer",
		slog.String("channel", channel),
		slog.String("name", name),
	)

	nextKiller.Start(db.Message{
		Channel:  chanState.Channel,
		Username: util.BotUsername,
		IsMod:    false,
		Text:     "",
	})

	return nil
}

func selectKillerWeighted(arr []killer.Killer, channel string) killer.Killer {
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
