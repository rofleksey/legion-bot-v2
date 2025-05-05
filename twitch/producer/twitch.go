package producer

import (
	"github.com/gempir/go-twitch-irc/v4"
	"github.com/samber/do"
	"legion-bot-v2/bot"
	"legion-bot-v2/config"
	"legion-bot-v2/db"
	"legion-bot-v2/twitch/twitch_api"
	"legion-bot-v2/util"
	"legion-bot-v2/util/taskq"
	"legion-bot-v2/util/timers"
	"log/slog"
	"os"
	"strings"
	"time"
)

var _ Producer = (*TwitchProducer)(nil)

type TwitchProducer struct {
	cfg            *config.Config
	timersInstance timers.Timers
	api            *twitch_api.TwitchApi
	database       db.DB
	botInstance    *bot.Bot
	queue          *taskq.Queue
}

func NewTwitchProducer(di *do.Injector) Producer {
	return &TwitchProducer{
		cfg:            do.MustInvoke[*config.Config](di),
		timersInstance: do.MustInvoke[timers.Timers](di),
		api:            do.MustInvoke[*twitch_api.TwitchApi](di),
		database:       do.MustInvoke[db.DB](di),
		botInstance:    do.MustInvoke[*bot.Bot](di),
		queue:          taskq.New(1, 1, 1),
	}
}

func (p *TwitchProducer) Shutdown() {
	p.api.IrcClient().Disconnect()
	p.queue.Shutdown()
}

func (p *TwitchProducer) Run() error {
	p.api.IrcClient().OnPrivateMessage(func(message twitch.PrivateMessage) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("OnPrivateMessage panic",
					slog.String("channel", message.Channel),
					slog.String("text", message.Message),
					slog.Any("error", err),
				)
			}
		}()

		username := strings.ToLower(message.User.Name)
		channel := strings.ReplaceAll(message.Channel, "#", "")
		text := strings.TrimSpace(message.Message)

		if username == util.BotUsername {
			return
		}

		modTagStr, _ := message.Tags["mod"]

		isMod := modTagStr == "1"

		slog.Debug("Message",
			slog.String("channel", channel),
			slog.String("username", username),
			slog.String("text", text),
			slog.Bool("isMod", isMod),
		)

		p.botInstance.HandleMessage(db.Message{
			ID:       message.ID,
			Channel:  channel,
			Username: username,
			IsMod:    isMod,
			Text:     text,
		})
	})

	p.api.IrcClient().OnUserNoticeMessage(func(message twitch.UserNoticeMessage) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("OnUserNoticeMessage panic",
					slog.String("channel", message.Channel),
					slog.String("text", message.Message),
					slog.Any("error", err),
				)
			}
		}()

		if message.MsgID == "raid" {
			channel := strings.ReplaceAll(message.Channel, "#", "")
			otherChannel := strings.ReplaceAll(message.MsgParams["msg-param-login"], "#", "")

			slog.Info("Incoming Raid",
				slog.String("channel", channel),
				slog.String("otherChannel", otherChannel),
			)

			p.botInstance.HandleIncomingRaid(channel, otherChannel)
		}
	})

	p.api.IrcClient().OnConnect(func() {
		slog.Info("Connected to IRC")
	})

	p.api.IrcClient().OnWhisperMessage(func(message twitch.WhisperMessage) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("OnWhisperMessage panic",
					slog.String("username", message.User.Name),
					slog.String("text", message.Message),
					slog.Any("error", err),
				)
			}
		}()

		username := strings.ToLower(message.User.Name)
		text := strings.TrimSpace(message.Message)

		if username == util.BotUsername {
			return
		}

		slog.Info("Whisper Message",
			slog.String("username", username),
			slog.String("text", text),
		)

		p.botInstance.HandleWhisper(username, text)
	})

	if os.Getenv("ENVIRONMENT") == "production" {
		channels := p.database.GetAllChannelNames()

		for _, channel := range channels {
			chanState := p.database.GetState(channel)
			if chanState.Settings.Disabled {
				continue
			}

			p.AddChannel(chanState.Channel)
			time.Sleep(time.Second)
		}
	}

	return p.api.IrcClient().Connect()
}

func (p *TwitchProducer) AddChannel(channel string) {
	p.queue.Enqueue(func() {
		p.api.IrcClient().Join(channel)
	})

	p.queue.Enqueue(func() {
		p.registerAllListeners(channel)
	})
}

func (p *TwitchProducer) RemoveChannel(channel string) {
	p.queue.Enqueue(func() {
		p.api.IrcClient().Depart(channel)
	})

	p.queue.Enqueue(func() {
		p.removeAllListeners(channel)
	})

	p.database.UpdateState(channel, func(state *db.ChannelState) {
		if state.Killer != "" {
			state.Killer = ""
			state.KillerState = nil
			state.Date = time.Now()
		}

		p.timersInstance.StopChannelTimers(channel)
	})
}
