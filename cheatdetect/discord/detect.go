package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/elliotchance/pie/v2"
	"legion-bot-v2/cheatdetect/common"
	"legion-bot-v2/util"
	"net/http"
	"strings"
	"time"
)

const BaseURL = "http://discord-detector:3000"

//const BaseURL = "http://localhost:3000"

var _ common.Detector = (*Detector)(nil)

type Detector struct {
	HTTPClient *http.Client
}

func New() *Detector {
	return &Detector{
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type Member struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	GuildName   string `json:"guild_name"`
}

func (d *Detector) Name() string {
	return "discord"
}

func (d *Detector) Detect(ctx context.Context, username string) ([]common.DetectedUser, error) {
	reqBody := strings.NewReader(fmt.Sprintf(`{"username": "%s"}`, username))
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/search_user", BaseURL),
		reqBody,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := d.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var members []Member
	if err := json.NewDecoder(resp.Body).Decode(&members); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	members = pie.Filter(members, func(member Member) bool {
		if strings.Contains(member.Username, util.BotOwner) {
			return false
		}

		return true
	})

	return pie.Map(members, func(m Member) common.DetectedUser {
		if m.DisplayName == "" {
			m.DisplayName = "?"
		}

		return common.DetectedUser{
			Username: fmt.Sprintf("%s (%s)", m.Username, m.DisplayName),
			Site:     m.GuildName,
		}
	}), nil
}
