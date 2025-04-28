package gpt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"legion-bot-v2/config"
	"net/http"
	"sync"
	"time"
)

const yandexUrl = "https://llm.api.cloud.yandex.net/foundationModels/v1/completion"
const ModelUri = "gpt://%s/yandexgpt-lite/rc"

var _ Gpt = (*YandexGpt)(nil)

type YandexGpt struct {
	cfg    *config.Config
	client *http.Client

	m        sync.Mutex
	iamToken string
}

func NewYandexGpt(cfg *config.Config) *YandexGpt {
	return &YandexGpt{
		cfg: cfg,
		client: &http.Client{
			Timeout: time.Second * 30,
		},
	}
}

func (g *YandexGpt) Process(ctx context.Context, prompt Prompt) (string, error) {
	messages := make([]YandexMessage, 0, 2)

	if prompt.SystemText != "" {
		messages = append(messages, YandexMessage{
			Role: "system",
			Text: prompt.SystemText,
		})
	}

	if prompt.Text != "" {
		messages = append(messages, YandexMessage{
			Role: "user",
			Text: prompt.Text,
		})
	}

	body := YandexBody{
		ModelUri: fmt.Sprintf(ModelUri, g.cfg.Yandex.FolderID),
		CompletionOptions: YandexCompletionOptions{
			MaxTokens:   500,
			Temperature: prompt.Temperature,
		},
		Messages: messages,
	}

	jsonBytes, err := json.Marshal(&body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal body: %w", err)
	}

	iamToken, err := g.getIAMToken(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get IAM token: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, yandexUrl, bytes.NewReader(jsonBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+iamToken)
	req.Header.Set("X-Folder-ID", g.cfg.Yandex.FolderID)

	res, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	bytez, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	jsonStr := string(bytez)

	var yRes YandexResponse

	if err = json.Unmarshal([]byte(jsonStr), &yRes); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(yRes.Result.Alternatives) == 0 {
		return "", fmt.Errorf("no alternatives found")
	}

	return yRes.Result.Alternatives[0].Message.Text, nil
}
