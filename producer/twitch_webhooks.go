package producer

import (
	"fmt"
	"github.com/nicklaw5/helix/v2"
	"legion-bot-v2/db"
	"log/slog"
)

func (p *TwitchProducer) registerAllListeners(channel string) {
	broadcasterResp, err := p.helixClient.GetUsers(&helix.UsersParams{
		Logins: []string{channel},
	})
	if err != nil {
		slog.Error("Failed to get channel user info for listeners",
			slog.String("channel", channel),
			slog.Any("error", err),
		)
		return
	}
	if len(broadcasterResp.Data.Users) == 0 {
		slog.Error("Failed to get channel user info for listeners",
			slog.String("channel", channel),
			slog.String("error", broadcasterResp.Error),
			slog.String("errorMsg", broadcasterResp.ErrorMessage),
		)
		return
	}

	broadcasterID := broadcasterResp.Data.Users[0].ID

	p.queue.Enqueue(func() {
		p.addOutgoingRaidsListener(channel, broadcasterID)
	})
	p.queue.Enqueue(func() {
		p.addStreamStartListener(channel, broadcasterID)
	})
}

func (p *TwitchProducer) addStreamStartListener(channel, broadcasterID string) {
	chanState := p.database.GetState(channel)
	streamStartId := chanState.Subs.StreamStart

	if streamStartId != "" {
		_, _ = p.appClient.RemoveEventSubSubscription(streamStartId)
	}

	resp, err := p.appClient.CreateEventSubSubscription(&helix.EventSubSubscription{
		Type:    "stream.online",
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: broadcasterID,
		},
		Transport: helix.EventSubTransport{
			Method:   "webhook",
			Callback: fmt.Sprintf("%s/api/webhook/stream/start", p.cfg.BaseURL),
			Secret:   p.cfg.Twitch.WebHookSecret,
		},
	})
	if err != nil {
		slog.Error("Failed to create event sub for stream start",
			slog.String("channel", channel),
			slog.Any("error", err),
		)
		return
	}
	if len(resp.Data.EventSubSubscriptions) == 0 {
		slog.Error("Failed to create event sub for stream start",
			slog.String("channel", channel),
			slog.String("error", resp.Error),
			slog.String("errorMsg", resp.ErrorMessage),
		)
		return
	}

	sub := resp.Data.EventSubSubscriptions[0]
	p.database.UpdateState(channel, func(state *db.ChannelState) {
		state.Subs.StreamStart = sub.ID
	})
	slog.Debug("Successfully created event sub for stream start",
		slog.String("channel", channel),
	)
}

func (p *TwitchProducer) addOutgoingRaidsListener(channel, broadcasterID string) {
	chanState := p.database.GetState(channel)
	raidSubId := chanState.Subs.RaidID

	if raidSubId != "" {
		_, _ = p.appClient.RemoveEventSubSubscription(raidSubId)
	}

	resp, err := p.appClient.CreateEventSubSubscription(&helix.EventSubSubscription{
		Type:    helix.EventSubTypeChannelRaid,
		Version: "1",
		Condition: helix.EventSubCondition{
			FromBroadcasterUserID: broadcasterID,
		},
		Transport: helix.EventSubTransport{
			Method:   "webhook",
			Callback: fmt.Sprintf("%s/api/webhook/raids", p.cfg.BaseURL),
			Secret:   p.cfg.Twitch.WebHookSecret,
		},
	})
	if err != nil {
		slog.Error("Failed to create event sub for raids",
			slog.String("channel", channel),
			slog.Any("error", err),
		)
		return
	}
	if len(resp.Data.EventSubSubscriptions) == 0 {
		slog.Error("Failed to create event sub for raids",
			slog.String("channel", channel),
			slog.String("error", resp.Error),
			slog.String("errorMsg", resp.ErrorMessage),
		)
		return
	}

	sub := resp.Data.EventSubSubscriptions[0]
	p.database.UpdateState(channel, func(state *db.ChannelState) {
		state.Subs.RaidID = sub.ID
	})
	slog.Debug("Successfully created event sub for outgoing raids",
		slog.String("channel", channel),
	)
}

func (p *TwitchProducer) removeAllListeners(channel string) {
	chanState := p.database.GetState(channel)

	if chanState.Subs.RaidID != "" {
		p.queue.Enqueue(func() {
			_, _ = p.appClient.RemoveEventSubSubscription(chanState.Subs.RaidID)
			slog.Debug("Removed event sub for raid subscription",
				slog.String("channel", channel),
			)
		})
	}

	if chanState.Subs.StreamStart != "" {
		p.queue.Enqueue(func() {
			_, _ = p.appClient.RemoveEventSubSubscription(chanState.Subs.StreamStart)
		})
		slog.Debug("Removed event sub for stream start subscription",
			slog.String("channel", channel),
		)
	}

	p.database.UpdateState(channel, func(state *db.ChannelState) {
		state.Subs.RaidID = ""
		state.Subs.StreamStart = ""
	})
}
