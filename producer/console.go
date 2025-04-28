package producer

import (
	"bufio"
	"fmt"
	"legion-bot-v2/bot"
	"legion-bot-v2/db"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"
)

var _ Producer = (*ConsoleProducer)(nil)

type ConsoleProducer struct {
	botInstance *bot.Bot

	m       sync.Mutex
	channel string
}

func NewConsoleProducer(botInstance *bot.Bot) *ConsoleProducer {
	return &ConsoleProducer{
		botInstance: botInstance,
	}
}

func (p *ConsoleProducer) Run() error {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()

		p.m.Lock()
		channel := p.channel
		p.m.Unlock()

		splitted := strings.SplitN(line, " ", 2)
		username := splitted[0]
		text := splitted[1]

		slog.Debug("Message",
			slog.String("channel", channel),
			slog.String("username", username),
			slog.String("text", text),
			slog.Bool("isMod", false),
		)

		p.botInstance.HandleMessage(db.Message{
			ID:       time.Now().String(),
			Channel:  channel,
			Username: username,
			IsMod:    false,
			Text:     text,
		})
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	return nil
}

func (p *ConsoleProducer) Stop() {

}

func (p *ConsoleProducer) AddChannel(channel string) {
	p.m.Lock()
	defer p.m.Unlock()

	p.channel = channel
}

func (p *ConsoleProducer) RemoveChannel(channel string) {

}
