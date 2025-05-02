package main

import (
	_ "embed"
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/samber/do"
	slogmulti "github.com/samber/slog-multi"
	slogtelegram "github.com/samber/slog-telegram/v2"
	"legion-bot-v2/api"
	"legion-bot-v2/bot"
	"legion-bot-v2/chat"
	"legion-bot-v2/cheatdetect"
	"legion-bot-v2/config"
	"legion-bot-v2/db"
	"legion-bot-v2/gpt"
	"legion-bot-v2/i18n"
	"legion-bot-v2/killer"
	"legion-bot-v2/killer/doctor"
	"legion-bot-v2/killer/dredge"
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
// display status
// display problems
// add more Info icons in the settings
// more descriptive killer powers in chat
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
// notes about players
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

	userAccessToken, err := util.FetchTwitchUserAccessToken(cfg)
	if err != nil {
		log.Fatalf("Failed to get Twitch access token: %v", err)
	}
	if os.Getenv("ENVIRONMENT") != "production" {
		slog.Debug("Got access token",
			slog.String("token", userAccessToken),
			slog.String("clientId", cfg.Twitch.ClientID),
		)
	}

	do.ProvideNamedValue(di, "userAccessToken", userAccessToken)

	ircClient, helixClient, err := util.InitTwitchClients(cfg, userAccessToken)
	if err != nil {
		log.Fatalf("Failed to init twitch clients: %v", err)
	}

	do.ProvideValue(di, ircClient)
	do.ProvideNamedValue(di, "helixClient", helixClient)

	appClient, err := util.InitAppTwitchClient(cfg, userAccessToken)
	if err != nil {
		log.Fatalf("Failed to init app twitch client: %v", err)
	}
	do.ProvideNamedValue(di, "appClient", appClient)

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
