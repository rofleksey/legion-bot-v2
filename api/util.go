package api

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"legion-bot-v2/api/dao"
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
