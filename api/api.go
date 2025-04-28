package api

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/jellydator/ttlcache/v3"
	"legion-bot-v2/config"
	"legion-bot-v2/db"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/twitch"
)

type Server struct {
	cfg          *config.Config
	oauth2Config oauth2.Config
	database     db.DB
	stateCache   *ttlcache.Cache[string, struct{}]
	mux          *http.ServeMux
}

type TwitchUser struct {
	ID              string `json:"id"`
	Login           string `json:"login"`
	DisplayName     string `json:"display_name"`
	ProfileImageURL string `json:"profile_image_url"`
	Email           string `json:"email"`
}

type JWTClaims struct {
	TwitchUser
	jwt.RegisteredClaims
}

func NewServer(cfg *config.Config, database db.DB) *Server {
	stateCache := ttlcache.New[string, struct{}](
		ttlcache.WithTTL[string, struct{}](30 * time.Minute),
	)

	server := Server{
		cfg: cfg,
		oauth2Config: oauth2.Config{
			ClientID:     cfg.Auth.ClientID,
			ClientSecret: cfg.Auth.ClientSecret,
			Endpoint:     twitch.Endpoint,
			RedirectURL:  cfg.Auth.RedirectURL,
			Scopes:       []string{"user:read:email"},
		},
		database:   database,
		stateCache: stateCache,
		mux:        http.NewServeMux(),
	}

	server.mux.HandleFunc("/api/auth/login", server.handleLogin)
	server.mux.HandleFunc("/api/auth/callback", server.handleCallback)
	server.mux.HandleFunc("/api/settings", server.handleSettings)
	server.mux.HandleFunc("/api/validate", server.handleValidateToken)
	server.mux.HandleFunc("/api/stats/{channel}", server.handleChannelStats)
	server.mux.HandleFunc("/api/stats/{channel}/{username}", server.handleUserStats)

	server.mux.HandleFunc("/stats/{channel}", server.handleStatsPage)
	server.mux.Handle("/", http.FileServer(http.Dir("./static")))

	return &server
}

func (s *Server) Run() error {
	slog.Info("Started server on port 8080")
	return http.ListenAndServe(":8080", languageMiddleware(s.mux))
}

func languageMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") ||
			strings.HasPrefix(r.URL.Path, "/ru") ||
			strings.HasPrefix(r.URL.Path, "/en") {
			next.ServeHTTP(w, r)
			return
		}

		if cookie, err := r.Cookie("user_lang"); err == nil {
			if cookie.Value == "ru" && !strings.HasPrefix(r.URL.Path, "/ru") {
				http.Redirect(w, r, "/ru"+r.URL.Path, http.StatusFound)
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		if r.URL.Path == "/" {
			acceptLang := r.Header.Get("Accept-Language")
			if strings.Contains(acceptLang, "ru") {
				http.Redirect(w, r, "/ru", http.StatusFound)
				return
			} else {
				http.Redirect(w, r, "/en", http.StatusFound)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleStatsPage(w http.ResponseWriter, r *http.Request) {
	channel := strings.TrimPrefix(r.URL.Path, "/stats/")
	if channel == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if s.database.GetState(channel).Date.IsZero() {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	http.ServeFile(w, r, filepath.Join("static", "html", "stats.html"))
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := generateRandomString(32)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	s.stateCache.Set(state, struct{}{}, ttlcache.DefaultTTL)

	authURL := s.oauth2Config.AuthCodeURL(state)

	json.NewEncoder(w).Encode(map[string]string{
		"authUrl": authURL,
		"state":   state,
	})
}

func (s *Server) handleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	if state == "" || code == "" {
		http.Error(w, "Missing state or code", http.StatusBadRequest)
		return
	}

	if !s.stateCache.Has(state) {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	token, err := s.oauth2Config.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	client := s.oauth2Config.Client(r.Context(), token)
	req, err := http.NewRequest(http.MethodGet, "https://api.twitch.tv/helix/users", nil)
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	req.Header.Set("Client-ID", s.cfg.Auth.ClientID)

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var result struct {
		Data []TwitchUser `json:"data"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		http.Error(w, "Failed to decode user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(result.Data) == 0 {
		http.Error(w, "No user data returned", http.StatusInternalServerError)
		return
	}

	user := result.Data[0]

	jwtToken, err := s.createJWTToken(user)
	if err != nil {
		http.Error(w, "Failed to create token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	clientRedirectUrl := url.URL{
		Path: "/",
		RawQuery: url.Values{
			"token": []string{jwtToken},
			"state": []string{state},
		}.Encode(),
	}

	http.Redirect(w, r, clientRedirectUrl.String(), http.StatusFound)
}

func (s *Server) createJWTToken(user TwitchUser) (string, error) {
	claims := JWTClaims{
		TwitchUser: user,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "twitch-auth-server",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.Auth.JwtSecret))
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getSettings(w, r)
	case http.MethodPost:
		s.updateSettings(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) getSettings(w http.ResponseWriter, r *http.Request) {
	claims, err := s.authenticateRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	state := s.database.GetState(claims.Login)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state.Settings)
}

func (s *Server) updateSettings(w http.ResponseWriter, r *http.Request) {
	claims, err := s.authenticateRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	log.Printf("User %s updating settings", claims.Login)

	var newSettings db.Settings
	if err := json.NewDecoder(r.Body).Decode(&newSettings); err != nil {
		http.Error(w, "Invalid settings data", http.StatusBadRequest)
		return
	}

	s.database.UpdateState(claims.Login, func(state *db.ChannelState) {
		state.Settings = newSettings
	})

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleValidateToken(w http.ResponseWriter, r *http.Request) {
	claims, err := s.authenticateRequest(r)
	if err != nil {
		http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(claims.TwitchUser)
}

func (s *Server) handleChannelStats(w http.ResponseWriter, r *http.Request) {
	channel := r.PathValue("channel")

	state := s.database.GetState(channel)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state.Stats)
}

func (s *Server) handleUserStats(w http.ResponseWriter, r *http.Request) {
	channel := r.PathValue("channel")
	username := r.PathValue("username")

	state := s.database.GetState(channel)

	user := state.UserMap[username]
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user.Stats)
}

func (s *Server) authenticateRequest(r *http.Request) (*JWTClaims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("missing authorization header")
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenStr == authHeader {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.Auth.JwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

func generateRandomString(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
