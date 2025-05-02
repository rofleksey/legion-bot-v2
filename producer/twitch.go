package producer

import (
	"github.com/gempir/go-twitch-irc/v4"
	"github.com/nicklaw5/helix/v2"
	"legion-bot-v2/bot"
	"legion-bot-v2/config"
	"legion-bot-v2/db"
	"legion-bot-v2/taskq"
	"legion-bot-v2/util"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"
)

var _ Producer = (*TwitchProducer)(nil)

type TwitchProducer struct {
	cfg             *config.Config
	userAccessToken string
	ircClient       *twitch.Client
	helixClient     *helix.Client
	appClient       *helix.Client
	database        db.DB
	botInstance     *bot.Bot
	queue           *taskq.Queue

	m                sync.Mutex
	websocketClients map[string]*TwitchWebSocketClient
}

func NewTwitchProducer(
	cfg *config.Config,
	userAccessToken string,
	ircClient *twitch.Client,
	helixClient *helix.Client,
	appClient *helix.Client,
	database db.DB,
	botInstance *bot.Bot,
) *TwitchProducer {
	return &TwitchProducer{
		cfg:              cfg,
		userAccessToken:  userAccessToken,
		ircClient:        ircClient,
		helixClient:      helixClient,
		appClient:        appClient,
		database:         database,
		botInstance:      botInstance,
		queue:            taskq.New(1, 1, 1),
		websocketClients: make(map[string]*TwitchWebSocketClient),
	}
}

func (p *TwitchProducer) Shutdown() {
	p.ircClient.Disconnect()
	p.queue.Shutdown()
}

func (p *TwitchProducer) Run() error {
	p.ircClient.OnPrivateMessage(func(message twitch.PrivateMessage) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("Error in Twitch Private Message:",
					slog.Any("error", err),
					slog.Any("text", message.Message),
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

	p.ircClient.OnUserNoticeMessage(func(message twitch.UserNoticeMessage) {
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

	p.ircClient.OnConnect(func() {
		slog.Info("Connected to IRC")
	})

	p.ircClient.OnWhisperMessage(func(message twitch.WhisperMessage) {
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

	return p.ircClient.Connect()
}

func (p *TwitchProducer) AddChannel(channel string) {
	p.queue.Enqueue(func() {
		p.ircClient.Join(channel)
	})

	p.queue.Enqueue(func() {
		p.registerAllListeners(channel)
	})
}

func (p *TwitchProducer) RemoveChannel(channel string) {
	p.queue.Enqueue(func() {
		p.ircClient.Depart(channel)
	})

	p.queue.Enqueue(func() {
		p.removeAllListeners(channel)
	})
}
