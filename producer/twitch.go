package producer

import (
	"fmt"
	"github.com/gempir/go-twitch-irc/v4"
	"github.com/nicklaw5/helix/v2"
	"legion-bot-v2/bot"
	"legion-bot-v2/config"
	"legion-bot-v2/db"
	"legion-bot-v2/util"
	"log/slog"
	"os"
	"strings"
)

var _ Producer = (*TwitchProducer)(nil)

type TwitchProducer struct {
	cfg         *config.Config
	ircClient   *twitch.Client
	helixClient *helix.Client
	appClient   *helix.Client
	database    db.DB
	botInstance *bot.Bot
}

func NewTwitchProducer(
	cfg *config.Config,
	ircClient *twitch.Client,
	helixClient *helix.Client,
	appClient *helix.Client,
	database db.DB,
	botInstance *bot.Bot,
) *TwitchProducer {
	return &TwitchProducer{
		cfg:         cfg,
		ircClient:   ircClient,
		helixClient: helixClient,
		appClient:   appClient,
		database:    database,
		botInstance: botInstance,
	}
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
		p.database.ReadAllStates(func(chanState *db.ChannelState) {
			if chanState.Settings.Disabled {
				return
			}

			p.AddChannel(chanState.Channel)
		})
	}

	return p.ircClient.Connect()
}

func (p *TwitchProducer) AddChannel(channel string) {
	p.ircClient.Join(channel)

	go p.tryAddOutgoingRaidsListener(channel)
}

func (p *TwitchProducer) tryAddOutgoingRaidsListener(channel string) {
	chanState := p.database.GetState(channel)
	raidSubId := chanState.Subs.RaidID

	broadcasterResp, err := p.helixClient.GetUsers(&helix.UsersParams{
		Logins: []string{channel},
	})
	if err != nil {
		slog.Error("Failed to get channel user info from helix for outgoing raids",
			slog.String("channel", channel),
			slog.Any("error", err),
		)
		return
	}
	if len(broadcasterResp.Data.Users) == 0 {
		slog.Error("Failed to get channel user info from helix for outgoing raids",
			slog.String("channel", channel),
			slog.String("error", broadcasterResp.Error),
			slog.String("errorMsg", broadcasterResp.ErrorMessage),
		)
		return
	}

	broadcasterID := broadcasterResp.Data.Users[0].ID

	if raidSubId != "" {
		_, _ = p.appClient.RemoveEventSubSubscription(raidSubId)
	}

	resp, err := p.appClient.CreateEventSubSubscription(&helix.EventSubSubscription{
		Type:    helix.EventSubTypeChannelRaid,
		Version: "1",
		Condition: helix.EventSubCondition{
			FromBroadcasterUserID: broadcasterID,
		},
		Transport: helix.EventSubTransport{
			Method:   "webhook",
			Callback: fmt.Sprintf("%s/api/webhook/raids", p.cfg.BaseURL),
			Secret:   p.cfg.Chat.WebHookSecret,
		},
	})
	if err != nil {
		slog.Error("Failed to create event sub for raids",
			slog.String("channel", channel),
			slog.Any("error", err),
		)
		return
	}
	if len(resp.Data.EventSubSubscriptions) == 0 {
		slog.Error("Failed to create event sub for raids",
			slog.String("channel", channel),
			slog.String("error", resp.Error),
			slog.String("errorMsg", resp.ErrorMessage),
		)
		return
	}

	sub := resp.Data.EventSubSubscriptions[0]
	p.database.UpdateState(channel, func(state *db.ChannelState) {
		state.Subs.RaidID = sub.ID
	})
}

func (p *TwitchProducer) tryRemoveOutgoingRaidsListener(channel string) {
	chanState := p.database.GetState(channel)
	raidSubId := chanState.Subs.RaidID

	if raidSubId != "" {
		_, _ = p.appClient.RemoveEventSubSubscription(raidSubId)
	}

	p.database.UpdateState(channel, func(state *db.ChannelState) {
		state.Subs.RaidID = ""
	})
}

func (p *TwitchProducer) RemoveChannel(channel string) {
	p.ircClient.Depart(channel)
	go p.tryRemoveOutgoingRaidsListener(channel)
}

func (p *TwitchProducer) Stop() {
	p.ircClient.Disconnect()
}
