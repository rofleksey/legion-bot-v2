package api

import (
	"encoding/json"
	"github.com/jellydator/ttlcache/v3"
	"legion-bot-v2/dao"
	"legion-bot-v2/db"
	"legion-bot-v2/util"
	"log"
	"net/http"
	"net/url"
	"time"
)

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
		Data []dao.TwitchUser `json:"data"`
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
	json.NewEncoder(w).Encode(dao.ResponseUser{
		ID:              claims.TwitchUser.ID,
		Login:           claims.TwitchUser.Login,
		DisplayName:     claims.TwitchUser.DisplayName,
		ProfileImageURL: claims.TwitchUser.ProfileImageURL,
		Email:           claims.TwitchUser.Email,
	})
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

func (s *Server) handleImport(w http.ResponseWriter, r *http.Request) {
	claims, err := s.authenticateRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	if claims.Login != util.BotOwner {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req dao.ImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	for _, legion := range req.Legions {
		if legion.Settings.Language == "" {
			legion.Settings.Language = "ru"
		}

		s.database.UpdateState(legion.Channel, func(state *db.ChannelState) {
			state.Channel = legion.Channel
			state.Date = time.Unix(0, legion.Date*int64(time.Millisecond))
			state.Stats = legion.Stats
			state.UserMap = legion.UserMap
			state.Settings = legion.Settings
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
