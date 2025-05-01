package producer

import (
	"fmt"
	"github.com/nicklaw5/helix/v2"
	"legion-bot-v2/db"
	"legion-bot-v2/util"
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

	botResp, err := p.helixClient.GetUsers(&helix.UsersParams{
		Logins: []string{util.BotUsername},
	})
	if err != nil {
		slog.Error("Failed to get bot user info for listeners",
			slog.String("channel", channel),
			slog.Any("error", err),
		)
		return
	}
	if len(botResp.Data.Users) == 0 {
		slog.Error("Failed to get channel user info for listeners",
			slog.String("channel", channel),
			slog.String("error", botResp.Error),
			slog.String("errorMsg", botResp.ErrorMessage),
		)
		return
	}

	botID := botResp.Data.Users[0].ID

	p.queue.Enqueue(func() {
		p.addOutgoingRaidsListener(channel, broadcasterID)
	})
	p.queue.Enqueue(func() {
		p.addGuestStarBeginListener(channel, broadcasterID, botID)
	})
	p.queue.Enqueue(func() {
		p.addGuestStarEndListener(channel, broadcasterID, botID)
	})
	p.queue.Enqueue(func() {
		p.addStreamStartListener(channel, broadcasterID)
	})
}

func (p *TwitchProducer) addGuestStarBeginListener(channel, broadcasterID, botID string) {
	chanState := p.database.GetState(channel)
	guestStarBeginId := chanState.Subs.GuestStarBegin

	if guestStarBeginId != "" {
		_, _ = p.appClient.RemoveEventSubSubscription(guestStarBeginId)
	}

	resp, err := p.appClient.CreateEventSubSubscription(&helix.EventSubSubscription{
		Type:    "channel.guest_star_session.begin",
		Version: "beta",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: broadcasterID,
			ModeratorUserID:   botID,
		},
		Transport: helix.EventSubTransport{
			Method:   "webhook",
			Callback: fmt.Sprintf("%s/api/webhook/guestStart/begin", p.cfg.BaseURL),
			Secret:   p.cfg.Twitch.WebHookSecret,
		},
	})
	if err != nil {
		slog.Error("Failed to create event sub for guest start begin",
			slog.String("channel", channel),
			slog.Any("error", err),
		)
		return
	}
	if len(resp.Data.EventSubSubscriptions) == 0 {
		slog.Error("Failed to create event sub for guest start begin",
			slog.String("channel", channel),
			slog.String("error", resp.Error),
			slog.String("errorMsg", resp.ErrorMessage),
		)
		return
	}

	sub := resp.Data.EventSubSubscriptions[0]
	p.database.UpdateState(channel, func(state *db.ChannelState) {
		state.Subs.GuestStarBegin = sub.ID
	})
}

func (p *TwitchProducer) addGuestStarEndListener(channel, broadcasterID, botID string) {
	chanState := p.database.GetState(channel)
	guestStarEndId := chanState.Subs.GuestStarEnd

	if guestStarEndId != "" {
		_, _ = p.appClient.RemoveEventSubSubscription(guestStarEndId)
	}

	resp, err := p.appClient.CreateEventSubSubscription(&helix.EventSubSubscription{
		Type:    "channel.guest_star_session.end",
		Version: "beta",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: broadcasterID,
			ModeratorUserID:   botID,
		},
		Transport: helix.EventSubTransport{
			Method:   "webhook",
			Callback: fmt.Sprintf("%s/api/webhook/guestStart/end", p.cfg.BaseURL),
			Secret:   p.cfg.Twitch.WebHookSecret,
		},
	})
	if err != nil {
		slog.Error("Failed to create event sub for guest start end",
			slog.String("channel", channel),
			slog.Any("error", err),
		)
		return
	}
	if len(resp.Data.EventSubSubscriptions) == 0 {
		slog.Error("Failed to create event sub for guest start end",
			slog.String("channel", channel),
			slog.String("error", resp.Error),
			slog.String("errorMsg", resp.ErrorMessage),
		)
		return
	}

	sub := resp.Data.EventSubSubscriptions[0]
	p.database.UpdateState(channel, func(state *db.ChannelState) {
		state.Subs.GuestStarEnd = sub.ID
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
}

func (p *TwitchProducer) removeAllListeners(channel string) {
	chanState := p.database.GetState(channel)

	if chanState.Subs.RaidID != "" {
		p.queue.Enqueue(func() {
			_, _ = p.appClient.RemoveEventSubSubscription(chanState.Subs.RaidID)
		})
	}

	if chanState.Subs.GuestStarBegin != "" {
		p.queue.Enqueue(func() {
			_, _ = p.appClient.RemoveEventSubSubscription(chanState.Subs.GuestStarBegin)
		})
	}

	if chanState.Subs.GuestStarEnd != "" {
		p.queue.Enqueue(func() {
			_, _ = p.appClient.RemoveEventSubSubscription(chanState.Subs.GuestStarEnd)
		})
	}

	if chanState.Subs.StreamStart != "" {
		p.queue.Enqueue(func() {
			_, _ = p.appClient.RemoveEventSubSubscription(chanState.Subs.StreamStart)
		})
	}

	p.database.UpdateState(channel, func(state *db.ChannelState) {
		state.Subs.RaidID = ""
		state.Subs.GuestStarBegin = ""
		state.Subs.GuestStarEnd = ""
		state.Subs.StreamStart = ""
	})
}
