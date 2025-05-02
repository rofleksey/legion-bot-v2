package api

import (
	"bytes"
	"encoding/json"
	"github.com/nicklaw5/helix/v2"
	"io"
	"legion-bot-v2/dao"
	"log/slog"
	"net/http"
	"strings"
)

func (s *Server) handleOutgoingRaid(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Failed to read outgoing raid body",
			slog.Any("error", err),
		)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
		return
	}
	defer r.Body.Close()

	if !helix.VerifyEventSubNotification(s.cfg.Twitch.WebHookSecret, r.Header, string(body)) {
		slog.Error("Invalid signature for outgoing raid")
		w.WriteHeader(http.StatusOK)
		return
	}

	var eventDao dao.EventSubNotification
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&eventDao)
	if err != nil {
		slog.Error("Failed to decode outgoing raid general body",
			slog.Any("error", err),
		)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
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
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
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

func (s *Server) handleStreamStart(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Failed to read stream start body",
			slog.Any("error", err),
		)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
		return
	}
	defer r.Body.Close()

	if !helix.VerifyEventSubNotification(s.cfg.Twitch.WebHookSecret, r.Header, string(body)) {
		slog.Error("Invalid signature for stream start")
		w.WriteHeader(http.StatusOK)
		return
	}

	var eventDao dao.EventSubNotification
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&eventDao)
	if err != nil {
		slog.Error("Failed to parse stream start general body",
			slog.Any("error", err),
		)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
		return
	}

	if eventDao.Challenge != "" {
		w.Write([]byte(eventDao.Challenge))
		return
	}

	var event dao.BroadcasterUserLoginEvent
	err = json.NewDecoder(bytes.NewReader(eventDao.Event)).Decode(&event)
	if err != nil {
		slog.Error("Failed to decode stream start body",
			slog.Any("error", err),
		)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
		return
	}

	channel := strings.ToLower(strings.ReplaceAll(event.BroadcasterUserLogin, "#", ""))

	slog.Error("Stream online",
		slog.String("channel", channel),
	)

	go s.bot.HandleStreamOnline(channel)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
