package gpt

import (
	"context"
	"github.com/stretchr/testify/require"
	"legion-bot-v2/config"
	"testing"
)

func TestYandex(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}

	g := NewYandexGpt(cfg)

	result, err := g.Process(context.Background(), Prompt{
		SystemText: "Исправь ошибки в тексте",
		Text:       "Привте Миръ",
	})
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, "Привет, Мир!", result)
}
