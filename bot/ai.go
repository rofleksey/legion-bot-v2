package bot

import (
	"context"
	"fmt"
	"legion-bot-v2/gpt"
	"regexp"
	"strings"
	"time"
)

func (b *Bot) GenericResponse(lang string, message string) (string, error) {
	promptText := b.GetLocalString(lang, "generic_response_prompt", map[string]string{})

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	result, err := b.Gpt.Process(ctx, gpt.Prompt{
		SystemText: promptText,
		Text:       message,
	})
	if err != nil {
		return "", fmt.Errorf("generic response prompt failed: %w", err)
	}

	re := regexp.MustCompile(`(?i)^RESULT\s+(.*)$`)

	matches := re.FindStringSubmatch(strings.TrimSpace(result))
	if len(matches) != 2 {
		return "", fmt.Errorf("generic response prompt failed: invalid result format: %s", result)
	}

	return strings.TrimSpace(matches[1]), nil
}
