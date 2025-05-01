package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jellydator/ttlcache/v3"
	"github.com/nicklaw5/helix/v2"
	"io"
	"legion-bot-v2/dao"
	"legion-bot-v2/db"
	"legion-bot-v2/util"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	slog.Info("Login attempt")

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

	slog.Info("Successful login",
		slog.String("login", user.Login),
	)

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

	slog.Info("Settings updated",
		slog.String("login", claims.TwitchUser.Login),
	)

	oldSettings := s.database.GetState(claims.Login).Settings

	var newSettings db.Settings
	if err := json.NewDecoder(r.Body).Decode(&newSettings); err != nil {
		http.Error(w, "Invalid settings data", http.StatusBadRequest)
		return
	}

	s.database.UpdateState(claims.TwitchUser.Login, func(state *db.ChannelState) {
		state.Settings = newSettings
	})

	if !oldSettings.Disabled && newSettings.Disabled {
		s.chatProducer.RemoveChannel(claims.TwitchUser.Login)
	} else if oldSettings.Disabled && !newSettings.Disabled {
		s.chatProducer.AddChannel(claims.TwitchUser.Login)
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleCheatDetect(w http.ResponseWriter, r *http.Request) {
	var reqBody dao.CheatDetectRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid settings data", http.StatusBadRequest)
		return
	}

	slog.Info("Detect request",
		slog.String("username", reqBody.Username),
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	res, err := s.cheatDetector.Detect(ctx, reqBody.Username)
	if err != nil {
		slog.Error("Failed to execute cheat detect request",
			slog.String("username", reqBody.Username),
			slog.Any("error", err),
		)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (s *Server) handleSummonKiller(w http.ResponseWriter, r *http.Request) {
	claims, err := s.authenticateRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	var reqBody dao.SummonKillerRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid killer data", http.StatusBadRequest)
		return
	}

	slog.Info("Summon killer request",
		slog.String("name", reqBody.Name),
	)

	if err = s.bot.StartSpecificKiller(claims.TwitchUser.Login, reqBody.Name); err != nil {
		slog.Error("Failed to summon killer manually",
			slog.String("channel", claims.TwitchUser.Login),
			slog.String("name", reqBody.Name),
			slog.Any("error", err),
		)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleUserList(w http.ResponseWriter, r *http.Request) {
	claims, err := s.authenticateRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	if claims.TwitchUser.Login != util.BotOwner {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var list []dao.AdminTwitchUser

	s.database.ReadAllStates(func(chanState *db.ChannelState) {
		list = append(list, dao.AdminTwitchUser{
			Login: chanState.Channel,
		})
	})

	if list == nil {
		list = []dao.AdminTwitchUser{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (s *Server) handleLoginAs(w http.ResponseWriter, r *http.Request) {
	claims, err := s.authenticateRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	if claims.TwitchUser.Login != util.BotOwner {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var reqBody dao.AdminTwitchUser
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid killer data", http.StatusBadRequest)
		return
	}

	twitchUser := dao.TwitchUser{
		Login:           reqBody.Login,
		DisplayName:     reqBody.Login,
		ProfileImageURL: fmt.Sprintf("%s/apple-touch-icon.png", s.cfg.BaseURL),
	}

	jwtToken, err := s.createJWTToken(twitchUser)
	if err != nil {
		http.Error(w, "Failed to create token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dao.AdminLoginResponse{
		Token: jwtToken,
		User: dao.ResponseTwitchUser{
			Login:           twitchUser.Login,
			DisplayName:     twitchUser.DisplayName,
			ProfileImageURL: twitchUser.ProfileImageURL,
		},
	})
}

func (s *Server) handleOutgoingRaid(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Failed to read outgoing raid body",
			slog.Any("error", err),
		)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if !helix.VerifyEventSubNotification(s.cfg.Chat.WebHookSecret, r.Header, string(body)) {
		slog.Error("Invalid signature for outgoing raid")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var eventDao dao.EventSubNotification
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&eventDao)
	if err != nil {
		slog.Error("Failed to decode outgoing raid general body",
			slog.Any("error", err),
		)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if eventDao.Challenge != "" {
		w.Write([]byte(eventDao.Challenge))
		return
	}

	var raidEvent helix.EventSubChannelRaidEvent

	err = json.NewDecoder(bytes.NewReader(eventDao.Event)).Decode(&raidEvent)
	if err != nil {
		slog.Error("Failed to decode outgoing raid body",
			slog.Any("error", err),
		)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fromChannel := strings.ReplaceAll(raidEvent.FromBroadcasterUserLogin, "#", "")
	toChannel := strings.ReplaceAll(raidEvent.ToBroadcasterUserLogin, "#", "")

	slog.Error("Outgoing raid",
		slog.String("from", fromChannel),
		slog.String("to", toChannel),
		slog.Int("viewers", raidEvent.Viewers),
	)

	go s.bot.HandleOutgoingRaid(fromChannel, toChannel)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (s *Server) handleValidateToken(w http.ResponseWriter, r *http.Request) {
	claims, err := s.authenticateRequest(r)
	if err != nil {
		http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dao.ResponseTwitchUser{
		Login:           claims.TwitchUser.Login,
		DisplayName:     claims.TwitchUser.DisplayName,
		ProfileImageURL: claims.TwitchUser.ProfileImageURL,
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
