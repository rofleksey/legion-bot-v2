package main

import (
	_ "embed"
	"fmt"
	"github.com/robfig/cron/v3"
	slogmulti "github.com/samber/slog-multi"
	slogtelegram "github.com/samber/slog-telegram/v2"
	"legion-bot-v2/api"
	"legion-bot-v2/bot"
	"legion-bot-v2/chat"
	"legion-bot-v2/config"
	"legion-bot-v2/db"
	"legion-bot-v2/gpt"
	"legion-bot-v2/i18n"
	"legion-bot-v2/killer"
	"legion-bot-v2/killer/doctor"
	"legion-bot-v2/killer/ghostface"
	"legion-bot-v2/killer/legion"
	"legion-bot-v2/killer/pinhead"
	"legion-bot-v2/producer"
	"legion-bot-v2/timers"
	"legion-bot-v2/util"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// TODO:
// greetings
// ai
// !spin
// !flashlight
// myers
// !clip
// !perk
// !addon
// voting
// flashlight blind
// chase clip with autoupload
// scoreboard autosave
// other streamers notification
// !chatgpt
// match timestamps
// dead hard notification
// ds notification
// !insult
// changelog
// changelog notification
// help page (!killer)
// current addons help (!addons) (!perks)
// chatbot ai
// monitor profile messages
// monitor stats changes
// notes about players
// steam related checks
// fmp/dbdtools checks discord
// played before
// gallery of killers
// documentation
// privacy policy
// terms

//go:embed banner.txt
var banner string

func main() {
	fmt.Fprintln(os.Stderr, banner)

	os.MkdirAll("data", 0750)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	logHandlers := []slog.Handler{
		slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}),
	}

	if os.Getenv("ENVIRONMENT") == "production" {
		logHandlers = append(logHandlers, slogtelegram.Option{
			Level:    slog.LevelInfo,
			Token:    cfg.Telegram.Token,
			Username: cfg.Telegram.ChatID,
		}.NewTelegramHandler())
	}

	multiHandler := slogmulti.Fanout(logHandlers...)
	slog.SetDefault(slog.New(multiHandler))

	slog.Info("Starting service...")

	c := cron.New()
	if _, err := c.AddFunc(util.DailyRestartCron, func() {
		log.Println("Executing daily restart")
		time.Sleep(time.Second * 1)
		os.Exit(1)
	}); err != nil {
		log.Fatalf("Failed to schedule daily restart: %v", err)
	}
	c.Start()

	accessToken, err := util.FetchTwitchAccessToken(cfg.Chat.RefreshToken)
	if err != nil {
		log.Fatalf("Failed to get Twitch access token: %v", err)
	}
	if os.Getenv("ENVIRONMENT") != "production" {
		slog.Debug("Got access token",
			slog.String("token", accessToken),
			slog.String("clientId", cfg.Chat.ClientID),
		)
	}

	ircClient, helixClient, err := util.InitTwitchClients(cfg.Chat.ClientID, accessToken)
	if err != nil {
		log.Fatalf("Failed to init twitch clients: %v", err)
	}

	var chatActions chat.Actions
	if os.Getenv("ENVIRONMENT") == "production" {
		chatActions = chat.NewTwitchActions(ircClient, helixClient)
	} else {
		slog.Debug("!!! Using debug chat actions")
		chatActions = &chat.ConsoleActions{}
	}

	database, err := db.NewDatabase("data/database.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	timerManager := timers.NewManager()

	localiser, err := i18n.NewLocaliser()
	if err != nil {
		log.Fatalf("Failed to initialize i18n: %v", err)
	}

	gptInstance := gpt.NewYandexGpt(cfg)

	killerMap := map[string]killer.Killer{
		"legion":    legion.New(database, chatActions, timerManager, localiser, gptInstance),
		"ghostface": ghostface.New(database, chatActions, timerManager, localiser, gptInstance),
		"doctor":    doctor.New(database, chatActions, timerManager, localiser, gptInstance),
		"pinhead":   pinhead.New(database, chatActions, timerManager, localiser, gptInstance),
	}

	botInstance := bot.NewBot(database, chatActions, timerManager, localiser, gptInstance, killerMap)
	botInstance.Init()

	//chatProducer := producer.NewTwitchProducer(ircClient, helixClient, database, botInstance)
	//if os.Getenv("ENVIRONMENT") != "production" {
	//	chatProducer.AddChannel("tru3ta1ent")
	//}

	chatProducer := producer.NewConsoleProducer(botInstance)
	chatProducer.AddChannel("rofleksey")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	go func() {
		<-sig
		chatProducer.Stop()
	}()

	slog.Debug("Starting server...")
	server := api.NewServer(cfg, database, chatProducer)
	go func() {
		if err := server.Run(); err != nil {
			slog.Error("Server error",
				slog.Any("error", err),
			)

			time.Sleep(time.Second)
			os.Exit(1)
		}
	}()

	slog.Debug("Connecting to twitch...")
	if err = chatProducer.Run(); err != nil {
		log.Fatalf("Chat listener failed: %v", err)
	}
}
