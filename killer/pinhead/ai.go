package pinhead

import (
	"context"
	"fmt"
	"github.com/elliotchance/pie/v2"
	"legion-bot-v2/gpt"
	"regexp"
	"strings"
	"time"
)

type GenerateWordResult struct {
	Topic string
	Word  string
}

func (p *Pinhead) GenerateWord(channel string) (GenerateWordResult, error) {
	chanState := p.GetState(channel)
	lang := chanState.Settings.Language

	rawTopics := strings.Split(strings.TrimSpace(strings.ToLower(chanState.Settings.Killers.Pinhead.Topics)), ",")
	topics := pie.Filter(pie.Map(rawTopics, func(t string) string {
		return strings.TrimSpace(t)
	}), func(s string) bool {
		return s != ""
	})

	if len(topics) == 0 {
		topics = []string{"города"}
	}

	promptText := p.GetLocalString(lang, "pinhead_generate_prompt", map[string]string{"TOPIC_LIST": strings.Join(topics, ",")})

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	result, err := p.Gpt.Process(ctx, gpt.Prompt{
		Text: promptText,
	})
	if err != nil {
		return GenerateWordResult{}, fmt.Errorf("word generation prompt failed: %w", err)
	}

	re := regexp.MustCompile(`(?i)^RESULT\s+(\w+)\s+(\w+)$`)

	matches := re.FindStringSubmatch(strings.TrimSpace(result))
	if len(matches) != 3 {
		return GenerateWordResult{}, fmt.Errorf("word generation prompt failed: invalid result format: %s", result)
	}

	topic := matches[1]
	word := matches[2]

	return GenerateWordResult{
		Topic: strings.ToLower(topic),
		Word:  strings.ToLower(word),
	}, nil
}

type GuessResult string

var (
	GuessResultYes       = GuessResult("yes")
	GuessResultNo        = GuessResult("no")
	GuessResultOK        = GuessResult("ok")
	GuessResultMaybe     = GuessResult("maybe")
	GuessResultIncorrect = GuessResult("incorrect")
)

func (p *Pinhead) GuessWord(lang, word, question string) (GuessResult, error) {
	promptText := p.GetLocalString(lang, "pinhead_guess_prompt", map[string]string{"THE_WORD": word})

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	result, err := p.Gpt.Process(ctx, gpt.Prompt{
		SystemText: promptText,
		Text:       question,
	})
	if err != nil {
		return GuessResultIncorrect, fmt.Errorf("guessing prompt failed: %w", err)
	}

	result = strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(result))), " ")

	switch result {
	case "ok":
		return GuessResultOK, nil
	case "ans y":
		return GuessResultYes, nil
	case "ans no":
		return GuessResultNo, nil
	case "maybe":
		return GuessResultMaybe, nil
	case "incorrect":
		return GuessResultIncorrect, nil
	default:
		return GuessResultIncorrect, fmt.Errorf("invalid guess result %s", result)
	}
}
