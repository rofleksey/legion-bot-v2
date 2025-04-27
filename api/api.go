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
}

type TwitchUser struct {
	ID          string `json:"id"`
	Login       string `json:"login"`
	DisplayName string `json:"display_name"`
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
	}

	http.HandleFunc("/api/auth/login", server.handleLogin)
	http.HandleFunc("/api/auth/callback", server.handleCallback)
	http.HandleFunc("/api/settings", server.handleSettings)
	http.HandleFunc("/api/validate", server.handleValidateToken)
	http.Handle("/", http.FileServer(http.Dir("./static")))

	return &server
}

func (s *Server) Run() error {
	slog.Info("Started server on port 8080")
	return http.ListenAndServe(":8080", nil)
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
	resp, err := client.Get("https://api.twitch.tv/helix/users")
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

	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": jwtToken,
		"user":  user,
	})
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
	return token.SignedString(s.cfg.Auth.JwtSecret)
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
		return s.cfg.Auth.JwtSecret, nil
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
