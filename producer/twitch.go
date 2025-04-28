package producer

import (
	"encoding/json"
	"fmt"
	"github.com/gempir/go-twitch-irc/v4"
	"github.com/nicklaw5/helix/v2"
	"legion-bot-v2/bot"
	"legion-bot-v2/config"
	"legion-bot-v2/db"
	"legion-bot-v2/util"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

var _ Producer = (*TwitchProducer)(nil)

type TwitchProducer struct {
	ircClient   *twitch.Client
	helixClient *helix.Client
	database    db.DB
	botInstance *bot.Bot
}

func NewTwitchProducer(cfg *config.Config, database db.DB, botInstance *bot.Bot) (*TwitchProducer, error) {
	accessToken, err := getTwitchAccessToken(cfg.Chat.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get Twitch access token: %w", err)
	}

	ircClient, helixClient, err := initTwitchClients(cfg.Chat.ClientID, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to init twitch clients: %w", err)
	}

	return &TwitchProducer{
		ircClient:   ircClient,
		helixClient: helixClient,
		database:    database,
		botInstance: botInstance,
	}, nil
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

	states := p.database.GetAllStates()
	for _, state := range states {
		p.AddChannel(state.Channel)
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

func getTwitchAccessToken(refreshToken string) (string, error) {
	client := &http.Client{
		Timeout: time.Second * 30,
	}

	req, err := http.NewRequest("GET", util.RefreshURL+"/"+refreshToken, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to refresh token: %s", resp.Status)
	}

	var result struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Token, nil
}

func initTwitchClients(clientID, accessToken string) (*twitch.Client, *helix.Client, error) {
	ircClient := twitch.NewClient(clientID, "oauth:"+accessToken)

	helixClient, err := helix.NewClient(&helix.Options{
		ClientID: clientID,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create helix client: %w", err)
	}

	helixClient.SetUserAccessToken(accessToken)

	return ircClient, helixClient, nil
}
