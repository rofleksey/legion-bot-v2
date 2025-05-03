package api

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/elliotchance/pie/v2"
	"github.com/golang-jwt/jwt/v5"
	"legion-bot-v2/api/dao"
	"legion-bot-v2/bot/killer"
	"legion-bot-v2/db"
	"net/http"
	"strings"
	"time"
)

type JWTClaims struct {
	dao.TwitchUser
	jwt.RegisteredClaims
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

func (s *Server) createJWTToken(user dao.TwitchUser) (string, error) {
	claims := JWTClaims{
		TwitchUser: user,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "twitch-auth-server",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.Auth.JwtSecret))
}

func generateRandomString(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (s *Server) formatChannelStatus(chanState db.ChannelState) dao.ChannelStatusResponse {
	lang := chanState.Settings.Language
	diff := time.Now().Sub(chanState.Date)
	generalKillerSettings := chanState.Settings.Killers.General

	if chanState.Settings.Disabled {
		return dao.ChannelStatusResponse{
			Status:   dao.ChannelStatusError,
			Title:    s.localiser.GetLocalString(lang, "channel_status_disabled", nil),
			Subtitle: s.localiser.GetLocalString(lang, "channel_status_disabled_subtitle", nil),
		}
	}

	if time.Now().Before(chanState.UserTimeout) {
		return dao.ChannelStatusResponse{
			Status:        dao.ChannelStatusError,
			Title:         s.localiser.GetLocalString(lang, "channel_status_user_timeout", nil),
			Subtitle:      s.localiser.GetLocalString(lang, "time_remaining_subtitle", nil),
			TimeRemaining: time.Until(chanState.UserTimeout),
		}
	}

	if chanState.Killer != "" {
		killerName := s.localiser.GetLocalString(lang, "killer_"+chanState.Killer, nil)

		var timeRemaining time.Duration

		k, _ := s.killerMap[chanState.Killer]
		if k != nil {
			timeRemaining = k.TimeRemaining(chanState.Channel)
		}

		return dao.ChannelStatusResponse{
			Status:        dao.ChannelStatusLoading,
			Title:         s.localiser.GetLocalString(lang, "channel_status_killer", map[string]string{"KILLER": killerName}),
			Subtitle:      s.localiser.GetLocalString(lang, "time_remaining_subtitle", nil),
			TimeRemaining: timeRemaining,
		}
	}

	killerList := pie.Filter(pie.Values(s.killerMap), func(k killer.Killer) bool {
		return k.Enabled(chanState.Channel)
	})
	if len(killerList) == 0 {
		return dao.ChannelStatusResponse{
			Status:   dao.ChannelStatusError,
			Title:    s.localiser.GetLocalString(lang, "channel_status_all_killers_disabled", nil),
			Subtitle: s.localiser.GetLocalString(lang, "channel_status_all_killers_disabled_title", nil),
		}
	}

	if diff <= generalKillerSettings.DelayBetweenKillers {
		return dao.ChannelStatusResponse{
			Status:        dao.ChannelStatusIdle,
			Title:         s.localiser.GetLocalString(lang, "channel_status_delay_killers", nil),
			Subtitle:      s.localiser.GetLocalString(lang, "time_remaining_subtitle", nil),
			TimeRemaining: generalKillerSettings.DelayBetweenKillers - diff,
		}
	}

	streamStartTime := s.bot.GetCachedStreamStartTime(chanState.Channel)

	var streamLength time.Duration
	if !streamStartTime.IsZero() {
		streamLength = time.Now().Sub(streamStartTime)
	}

	if streamLength <= generalKillerSettings.DelayAtTheStreamStart {
		return dao.ChannelStatusResponse{
			Status:        dao.ChannelStatusIdle,
			Title:         s.localiser.GetLocalString(lang, "channel_status_delay_stream_start", nil),
			Subtitle:      s.localiser.GetLocalString(lang, "time_remaining_subtitle", nil),
			TimeRemaining: generalKillerSettings.DelayAtTheStreamStart - streamLength,
		}
	}

	viewerCount := s.bot.GetCachedViewerCount(chanState.Channel)
	if viewerCount < generalKillerSettings.MinNumberOfViewers {
		needCount := generalKillerSettings.MinNumberOfViewers - viewerCount
		return dao.ChannelStatusResponse{
			Status:        dao.ChannelStatusIdle,
			Title:         s.localiser.GetLocalString(lang, "channel_status_not_enough_viewers", nil),
			Subtitle:      s.localiser.GetLocalString(lang, "channel_status_not_enough_viewers_subtitle", map[string]string{"COUNT": fmt.Sprint(needCount)}),
			TimeRemaining: generalKillerSettings.DelayAtTheStreamStart - streamLength,
		}
	}

	return dao.ChannelStatusResponse{
		Status:   dao.ChannelStatusSuccess,
		Title:    s.localiser.GetLocalString(lang, "channel_status_success", nil),
		Subtitle: s.localiser.GetLocalString(lang, "channel_status_success_subtitle", nil),
	}
}
