package producer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/exp/slog"
)

type TwitchWebSocketClient struct {
	conn             *websocket.Conn
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
	messageChan      chan TwitchEventMessage
	broadcasterID    string
	moderatorID      string
	clientID         string
	userAccessToken  string
	websocketURL     string
	reconnectTimeout time.Duration
	sessionID        string
}

type TwitchEventMessage struct {
	Metadata struct {
		MessageType string `json:"message_type"`
	} `json:"metadata"`
	Payload struct {
		Subscription struct {
			Type    string `json:"type"`
			Version string `json:"version"`
		} `json:"subscription"`
		Event struct {
			BroadcasterUserLogin string `json:"broadcaster_user_login"`
		} `json:"event"`
	} `json:"payload"`
}

func NewTwitchEventSubClient(broadcasterID, moderatorID, clientID, authToken string) *TwitchWebSocketClient {
	return &TwitchWebSocketClient{
		broadcasterID:    broadcasterID,
		moderatorID:      moderatorID,
		clientID:         clientID,
		userAccessToken:  authToken,
		websocketURL:     "wss://eventsub.wss.twitch.tv/ws",
		reconnectTimeout: 5 * time.Second,
		messageChan:      make(chan TwitchEventMessage, 100),
	}
}

func (t *TwitchWebSocketClient) Connect() error {
	ctx, cancel := context.WithCancel(context.Background())
	t.ctx = ctx
	t.cancel = cancel

	if err := t.connectWebSocket(); err != nil {
		slog.Error("WebSocket connection failed", "error", err)
		return fmt.Errorf("websocket connection failed: %w", err)
	}

	sessionID, err := t.waitForSessionID()
	if err != nil {
		slog.Error("Failed to get session ID", "error", err)
		return fmt.Errorf("session ID error: %w", err)
	}
	t.sessionID = sessionID

	if err := t.registerSubscriptions(); err != nil {
		slog.Error("Failed to register subscriptions", "error", err)
		return fmt.Errorf("subscription registration error: %w", err)
	}

	t.wg.Add(1)
	go t.readMessages()

	slog.Info("Twitch WebSocket client connected",
		"broadcasterID", t.broadcasterID,
		"moderatorID", t.moderatorID,
		"sessionID", t.sessionID,
	)

	return nil
}

func (t *TwitchWebSocketClient) connectWebSocket() error {
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(t.websocketURL, http.Header{})
	if err != nil {
		return err
	}
	t.conn = conn
	return nil
}

func (t *TwitchWebSocketClient) waitForSessionID() (string, error) {
	_, message, err := t.conn.ReadMessage()
	if err != nil {
		return "", err
	}

	var welcomeMsg struct {
		Metadata struct {
			MessageType string `json:"message_type"`
		} `json:"metadata"`
		Payload struct {
			Session struct {
				ID string `json:"id"`
			} `json:"session"`
		} `json:"payload"`
	}

	if err := json.Unmarshal(message, &welcomeMsg); err != nil {
		return "", fmt.Errorf("failed to unmarshal welcome message: %w", err)
	}

	if welcomeMsg.Metadata.MessageType != "session_welcome" {
		return "", errors.New("expected session_welcome message")
	}

	return welcomeMsg.Payload.Session.ID, nil
}

func (t *TwitchWebSocketClient) registerSubscriptions() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	events := []struct {
		Type    string
		Version string
	}{
		{"channel.guest_star_session.begin", "beta"},
		{"channel.guest_star_session.end", "beta"},
	}

	for _, event := range events {
		reqBody := map[string]interface{}{
			"type":    event.Type,
			"version": event.Version,
			"condition": map[string]string{
				"broadcaster_user_id": t.broadcasterID,
				"moderator_user_id":   t.moderatorID,
			},
			"transport": map[string]string{
				"method":     "websocket",
				"session_id": t.sessionID,
			},
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", "https://api.twitch.tv/helix/eventsub/subscriptions", bytes.NewBuffer(jsonBody))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+t.userAccessToken)
		req.Header.Set("Client-Id", t.clientID)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusAccepted {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		slog.Debug("Registered Twitch EventSub subscription",
			"type", event.Type,
			"version", event.Version,
			"broadcasterID", t.broadcasterID,
			"moderatorID", t.moderatorID,
		)
	}

	return nil
}

func (t *TwitchWebSocketClient) readMessages() {
	defer t.wg.Done()
	defer slog.Info("Stopped reading WebSocket messages")

	for {
		select {
		case <-t.ctx.Done():
			slog.Info("WebSocket message reader stopped by context")
			return
		default:
			_, message, err := t.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
					slog.Error("WebSocket read error",
						"error", err,
						"sessionID", t.sessionID,
					)
					t.reconnect()
				}
				return
			}

			var msg TwitchEventMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				slog.Error("Failed to unmarshal WebSocket message",
					"error", err,
					"message", string(message),
				)
				continue
			}

			switch msg.Metadata.MessageType {
			case "notification":
				select {
				case t.messageChan <- msg:
					slog.Debug("Received Twitch event",
						"type", msg.Payload.Subscription.Type,
						"broadcaster", msg.Payload.Event.BroadcasterUserLogin,
					)
				case <-t.ctx.Done():
					return
				}
			case "session_keepalive":
				slog.Debug("Received WebSocket keepalive")
			case "session_reconnect":
				slog.Info("Received WebSocket reconnect request")
				t.handleReconnect(message)
			default:
				slog.Warn("Received unknown message type",
					"type", msg.Metadata.MessageType,
				)
			}
		}
	}
}

func (t *TwitchWebSocketClient) handleReconnect(message []byte) {
	var reconnectMsg struct {
		Payload struct {
			Session struct {
				ReconnectURL string `json:"reconnect_url"`
			} `json:"session"`
		} `json:"payload"`
	}

	if err := json.Unmarshal(message, &reconnectMsg); err != nil {
		slog.Error("Failed to parse reconnect message",
			"error", err,
			"message", string(message),
		)
		return
	}

	if reconnectMsg.Payload.Session.ReconnectURL != "" {
		t.websocketURL = reconnectMsg.Payload.Session.ReconnectURL
		slog.Debug("Reconnecting WebSocket",
			"newURL", t.websocketURL,
			"sessionID", t.sessionID,
		)
		t.reconnect()
	}
}

func (t *TwitchWebSocketClient) reconnect() {
	if t.conn != nil {
		if err := t.conn.Close(); err != nil {
			slog.Error("Failed to close WebSocket during reconnect",
				"error", err,
				"sessionID", t.sessionID,
			)
		}
	}

	retryCount := 0
	maxRetries := 5

	for {
		select {
		case <-t.ctx.Done():
			slog.Debug("Reconnect attempt canceled by context")
			return
		default:
			if retryCount >= maxRetries {
				slog.Error("Max reconnect attempts reached",
					"attempts", retryCount,
					"sessionID", t.sessionID,
				)
				return
			}

			err := t.Connect()
			if err == nil {
				slog.Debug("Successfully reconnected WebSocket",
					"attempts", retryCount+1,
					"sessionID", t.sessionID,
				)
				return
			}

			retryCount++
			slog.Warn("WebSocket reconnect failed",
				"attempt", retryCount,
				"maxAttempts", maxRetries,
				"error", err,
				"retryIn", t.reconnectTimeout,
				"sessionID", t.sessionID,
			)
			time.Sleep(t.reconnectTimeout)
		}
	}
}

func (t *TwitchWebSocketClient) Messages() <-chan TwitchEventMessage {
	return t.messageChan
}

func (t *TwitchWebSocketClient) Close() error {
	if t.cancel != nil {
		t.cancel()
	}

	var closeErr error
	if t.conn != nil {
		slog.Debug("Closing WebSocket connection",
			"sessionID", t.sessionID,
		)

		if err := t.conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		); err != nil {
			slog.Error("Failed to send WebSocket close message",
				"error", err,
				"sessionID", t.sessionID,
			)
			closeErr = err
		}

		if err := t.conn.Close(); err != nil {
			slog.Error("Failed to close WebSocket",
				"error", err,
				"sessionID", t.sessionID,
			)
			if closeErr == nil {
				closeErr = fmt.Errorf("websocket close error: %w", err)
			}
		}
	}

	t.wg.Wait()
	close(t.messageChan)
	slog.Debug("WebSocket client fully shutdown",
		"sessionID", t.sessionID,
	)

	return closeErr
}

func StartWebsocketClient(
	broadcasterID string,
	moderatorID string,
	clientID string,
	authToken string,
	callback func(event, channel string),
) (*TwitchWebSocketClient, error) {
	client := NewTwitchEventSubClient(broadcasterID, moderatorID, clientID, authToken)
	if err := client.Connect(); err != nil {
		slog.Error("Failed to start WebSocket client",
			"broadcasterID", broadcasterID,
			"moderatorID", moderatorID,
			"error", err,
		)
		return nil, fmt.Errorf("websocket connect error: %w", err)
	}

	go func() {
		for msg := range client.Messages() {
			callback(msg.Payload.Subscription.Type, msg.Payload.Event.BroadcasterUserLogin)
		}
		slog.Debug("Message handler goroutine exited",
			"broadcasterID", broadcasterID,
			"moderatorID", moderatorID,
		)
	}()

	return client, nil
}
