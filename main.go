package main

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/samber/do"
	slogmulti "github.com/samber/slog-multi"
	slogtelegram "github.com/samber/slog-telegram/v2"
	"legion-bot-v2/api"
	"legion-bot-v2/bot"
	"legion-bot-v2/bot/i18n"
	"legion-bot-v2/bot/killer"
	"legion-bot-v2/bot/killer/doctor"
	"legion-bot-v2/bot/killer/dredge"
	"legion-bot-v2/bot/killer/ghostface"
	"legion-bot-v2/bot/killer/legion"
	"legion-bot-v2/bot/killer/pinhead"
	"legion-bot-v2/cheatdetect"
	"legion-bot-v2/config"
	"legion-bot-v2/db"
	"legion-bot-v2/gpt"
	"legion-bot-v2/steam"
	"legion-bot-v2/twitch/chat"
	"legion-bot-v2/twitch/producer"
	"legion-bot-v2/twitch/twitch_api"
	"legion-bot-v2/util"
	"legion-bot-v2/util/timers"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// TODO:
// display problems
// add more Info icons in the settings
// greetings
// !clip
// improve chatbot ai
// documentation
// killer specific help page (!killer) + generic (!legionbot) help page

// TODO: potential bugs
// check if messages are being processed sequentially

// TODO: these require link to a steam profile:
// monitor profile messages
// monitor stats changes

// TODO: these require the stream monitoring:
// rewind the whole stream
// chase clip with autoupload
// scoreboard autosave
// other streamers notification
// match timestamps
// decisive notification
// flashlight blind
// notes about streamers?
// steam related checks
// fmp/dbdtools checks discord
// played before
// voting

// TODO: these require perks/addons database:
// !perk
// !addon
// other streamers perks
// current addons help (!addons) (!perks)

// TODO: killer ideas:
// trapper - user gets into a trap by saying a word from a blacklist, others need to !untrap them or they get hooked
// dracula - reacts on caps keys, they get a hit
// pig - player with the most chat lines gets a headtrap - chat votes whether they explode or not
// pyramid head
// myers

// TODO: these are very minor but require a lot of pain:
// privacy policy
// terms
// !spin
// !flashlight
// gallery of killers
// changelog
// changelog notification

//go:embed banner.txt
var banner string

func main() {
	fmt.Fprintln(os.Stderr, banner)

	os.MkdirAll("data", 0750)

	di := do.New()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	do.ProvideValue(di, cfg)

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

	do.Provide(di, twitch_api.NewTwitchApi)
	go do.MustInvoke[*twitch_api.TwitchApi](di).Run(context.Background())

	var chatActions chat.Actions
	if os.Getenv("ENVIRONMENT") == "production" {
		chatActions = chat.NewTwitchActions(di)
	} else {
		slog.Debug("!!! Using debug chat actions")
		chatActions = &chat.ConsoleActions{}
	}
	defer chatActions.Shutdown()

	do.ProvideValue(di, chatActions)

	database, err := db.NewDatabase("data/database.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	do.ProvideValue(di, database)

	timerManager := timers.NewManager()
	do.ProvideValue(di, timerManager)

	localiser, err := i18n.NewLocaliser()
	if err != nil {
		log.Fatalf("Failed to initialize i18n: %v", err)
	}
	do.ProvideValue(di, localiser)

	gptInstance := gpt.NewYandexGpt(cfg)
	do.ProvideValue(di, gptInstance)

	killerMap := map[string]killer.Killer{
		"legion":    legion.New(di),
		"ghostface": ghostface.New(di),
		"doctor":    doctor.New(di),
		"pinhead":   pinhead.New(di),
		"dredge":    dredge.New(di),
	}
	do.ProvideValue(di, killerMap)

	botInstance := bot.NewBot(di)
	botInstance.Init()
	do.ProvideValue(di, botInstance)

	chatProducer := producer.NewTwitchProducer(di)
	defer chatProducer.Shutdown()
	do.ProvideValue(di, chatProducer)

	if os.Getenv("ENVIRONMENT") != "production" {
		chatProducer.AddChannel("dbdleague")
	}

	//chatProducer := producer.NewConsoleProducer(botInstance)
	//chatProducer.AddChannel("rofleksey")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	go func() {
		<-sig
		slog.Info("Shutting down...")
		chatProducer.Shutdown()
	}()

	cheatDetector := cheatdetect.NewDetector()
	do.ProvideValue(di, cheatDetector)

	steamClient, err := steam.NewClient(di)
	if err != nil {
		log.Fatalf("Failed to initialize steam client: %v", err)
	}
	go steamClient.Run(context.Background())
	do.ProvideValue(di, steamClient)

	slog.Debug("Starting server...")
	server := api.NewServer(di)
	do.ProvideValue(di, server)
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
