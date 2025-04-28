package producer

import (
	"github.com/gempir/go-twitch-irc/v4"
	"github.com/nicklaw5/helix/v2"
	"legion-bot-v2/bot"
	"legion-bot-v2/db"
	"legion-bot-v2/util"
	"log/slog"
	"os"
	"strings"
)

var _ Producer = (*TwitchProducer)(nil)

type TwitchProducer struct {
	ircClient   *twitch.Client
	helixClient *helix.Client
	database    db.DB
	botInstance *bot.Bot
}

func NewTwitchProducer(
	ircClient *twitch.Client,
	helixClient *helix.Client,
	database db.DB,
	botInstance *bot.Bot,
) *TwitchProducer {
	return &TwitchProducer{
		ircClient:   ircClient,
		helixClient: helixClient,
		database:    database,
		botInstance: botInstance,
	}
}

func (p *TwitchProducer) Run() error {
	p.ircClient.OnPrivateMessage(func(message twitch.PrivateMessage) {
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
			Channel:  channel,
			Username: username,
			IsMod:    isMod,
			Text:     text,
		})
	})
	p.ircClient.OnConnect(func() {
		slog.Info("Connected to IRC")
	})

	if os.Getenv("ENVIRONMENT") == "production" {
		states := p.database.GetAllStates()
		for _, state := range states {
			p.AddChannel(state.Channel)
		}
	}

	return p.ircClient.Connect()
}

func (p *TwitchProducer) AddChannel(channel string) {
	slog.Info("Channel added to chat producer",
		slog.String("channel", channel),
	)
	p.ircClient.Join(channel)
}

func (p *TwitchProducer) RemoveChannel(channel string) {
	slog.Info("Channel removed from chat producer",
		slog.String("channel", channel),
	)
	p.ircClient.Depart(channel)
}

func (p *TwitchProducer) Stop() {
	p.ircClient.Disconnect()
}
