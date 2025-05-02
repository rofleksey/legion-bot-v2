package api

import (
	"github.com/jellydator/ttlcache/v3"
	"legion-bot-v2/bot"
	"legion-bot-v2/cheatdetect"
	"legion-bot-v2/config"
	"legion-bot-v2/db"
	"legion-bot-v2/producer"
	"legion-bot-v2/util"
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/twitch"
)

type Server struct {
	cfg           *config.Config
	oauth2Config  oauth2.Config
	bot           *bot.Bot
	database      db.DB
	chatProducer  producer.Producer
	cheatDetector *cheatdetect.Detector
	stateCache    *ttlcache.Cache[string, struct{}]
	mux           *http.ServeMux
}

func NewServer(
	cfg *config.Config,
	bot *bot.Bot,
	database db.DB,
	chatProducer producer.Producer,
	cheatDetector *cheatdetect.Detector,
) *Server {
	stateCache := ttlcache.New[string, struct{}](
		ttlcache.WithTTL[string, struct{}](30 * time.Minute),
	)

	server := Server{
		cfg: cfg,
		bot: bot,
		oauth2Config: oauth2.Config{
			ClientID:     cfg.Twitch.ClientID,
			ClientSecret: cfg.Twitch.ClientSecret,
			Endpoint:     twitch.Endpoint,
			RedirectURL:  cfg.Twitch.RedirectURL,
			Scopes:       []string{},
		},
		database:      database,
		chatProducer:  chatProducer,
		cheatDetector: cheatDetector,
		stateCache:    stateCache,
		mux:           http.NewServeMux(),
	}

	server.mux.HandleFunc("/api/auth/login", server.handleLogin)
	server.mux.HandleFunc("/api/auth/callback", server.handleCallback)
	server.mux.HandleFunc("/api/validate", server.handleValidateToken)

	server.mux.HandleFunc("/api/settings", server.handleSettings)
	server.mux.HandleFunc("/api/channelState", server.handleChannelState)

	server.mux.HandleFunc("/api/stats/{channel}", server.handleChannelStats)
	server.mux.HandleFunc("/api/stats/{channel}/{username}", server.handleUserStats)
	server.mux.HandleFunc("/api/summonKiller", server.handleSummonKiller)

	server.mux.HandleFunc("/api/admin/users", server.handleUserList)
	server.mux.HandleFunc("/api/admin/loginAs", server.handleLoginAs)

	server.mux.HandleFunc("/api/webhook/raids", server.handleOutgoingRaid)
	server.mux.HandleFunc("/api/webhook/stream/start", server.handleStreamStart)

	server.mux.HandleFunc("/api/cheatDetect", server.handleCheatDetect)

	fs := http.FileServer(http.Dir("./frontend/dist"))
	cacheFS := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if util.IsStaticAsset(r.URL.Path) {
			w.Header().Set("Cache-Control", "public, max-age=31536000")
		}
		fs.ServeHTTP(w, r)
	})
	server.mux.Handle("/", cacheFS)

	return &server
}

func (s *Server) Run() error {
	slog.Debug("Started server on port 8080")
	return http.ListenAndServe(":8080", recoveryMiddleware(s.mux))
}
