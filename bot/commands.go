package bot

import (
	"fmt"
	"legion-bot-v2/db"
	"legion-bot-v2/util"
	"log/slog"
	"strings"
	"time"
)

func (b *Bot) HandleCommands(userMsg db.Message) bool {
	chanState := b.GetState(userMsg.Channel)
	lang := chanState.Settings.Language
	user := chanState.UserMap[userMsg.Username]

	switch {
	case strings.HasPrefix(userMsg.Text, "!legiontimeout") && userMsg.Username == userMsg.Channel:
		timeStr := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(userMsg.Text, "!legiontimeout")))

		duration, err := time.ParseDuration(timeStr)
		if err != nil {
			b.SendMessage(userMsg.Channel, fmt.Sprintf("Error: %v", err))
			return true
		}

		timeoutTime := time.Now().Add(duration)

		b.UpdateState(userMsg.Channel, func(chanState *db.ChannelState) {
			chanState.UserTimeout = timeoutTime
		})

		b.SendMessage(userMsg.Channel, fmt.Sprintf("Timeout till %v", timeoutTime.String()))

		b.UpdateState(userMsg.Channel, func(state *db.ChannelState) {
			if state.Killer != "" {
				state.Killer = ""
				state.KillerState = nil
				state.Date = time.Now()
			}

			b.StopChannelTimers(userMsg.Channel)
		})

		return true

	case strings.HasPrefix(userMsg.Text, "!hp"):
		otherUsername := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(strings.ReplaceAll(userMsg.Text, "@", ""), "!hp")))
		if otherUsername == "" {
			otherUsername = userMsg.Username
		}

		otherUser, userExists := chanState.UserMap[otherUsername]
		if !userExists {
			otherUser = db.NewUser()
			b.UpdateState(userMsg.Channel, func(chanState *db.ChannelState) {
				chanState.UserMap[otherUsername] = otherUser
			})
		}

		var msg string
		switch otherUser.Health {
		case "hooked":
			msg = b.GetLocalString(lang, "hooked", map[string]string{"USERNAME": otherUsername})
		case "deep_wound":
			msg = b.GetLocalString(lang, "deep_wound", map[string]string{"USERNAME": otherUsername})
		case "injured":
			msg = b.GetLocalString(lang, "injured", map[string]string{"USERNAME": otherUsername})
		case "dead":
			msg = b.GetLocalString(lang, "dead", map[string]string{"USERNAME": otherUsername})
		case "healthy":
			msg = b.GetLocalString(lang, "healthy", map[string]string{"USERNAME": otherUsername})
		default:
			return true
		}

		b.SendMessage(userMsg.Channel, msg)

		return true

	case strings.HasPrefix(userMsg.Text, "!unhook"):
		otherUsername := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(strings.ReplaceAll(userMsg.Text, "@", ""), "!unhook")))
		if otherUsername == "" {
			otherUsername = userMsg.Username
		}

		otherUser, userExists := chanState.UserMap[otherUsername]
		if !userExists {
			otherUser = db.NewUser()
			b.UpdateState(userMsg.Channel, func(chanState *db.ChannelState) {
				chanState.UserMap[otherUsername] = otherUser
			})
		}

		if otherUsername == userMsg.Username {
			msg := b.GetLocalString(lang, "cant_unhook_self", map[string]string{"USERNAME": otherUsername})
			b.SendMessage(userMsg.Channel, msg)

			return true
		}

		if otherUser.Health != "hooked" {
			msg := b.GetLocalString(lang, "not_hooked", map[string]string{"USERNAME": otherUsername})
			b.SendMessage(userMsg.Channel, msg)

			return true
		}

		b.UpdateState(userMsg.Channel, func(chanState *db.ChannelState) {
			chanState.UserMap[otherUsername].Health = "healthy"

			chanState.UserMap[userMsg.Username].Stats["unhooks"]++
		})

		b.UnbanUser(userMsg.Channel, otherUsername)
		b.StopTimer(userMsg.Channel, otherUsername)

		msg := b.GetLocalString(lang, "on_unhooked", map[string]string{"USERNAME": otherUsername})
		b.SendMessage(userMsg.Channel, msg)

		return true

	case strings.HasPrefix(userMsg.Text, "!heal"):
		otherUsername := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(strings.ReplaceAll(userMsg.Text, "@", ""), "!heal")))
		if otherUsername == "" {
			otherUsername = userMsg.Username
		}

		otherUser, userExists := chanState.UserMap[otherUsername]
		if !userExists {
			otherUser = db.NewUser()
			b.UpdateState(userMsg.Channel, func(chanState *db.ChannelState) {
				chanState.UserMap[otherUsername] = otherUser
			})
		}

		if otherUsername == userMsg.Username {
			msg := b.GetLocalString(lang, "cant_heal_self", map[string]string{"USERNAME": otherUsername})
			b.SendMessage(userMsg.Channel, msg)
			return true
		}

		if user.Health == "hooked" || user.Health == "dead" {
			msg := b.GetLocalString(lang, "cant_do_rn", map[string]string{"USERNAME": otherUsername})
			b.SendMessage(userMsg.Channel, msg)
			return true
		}

		if otherUser.Health == "hooked" {
			msg := b.GetLocalString(lang, "hooked", map[string]string{"USERNAME": otherUsername})
			b.SendMessage(userMsg.Channel, msg)
			return true
		}

		if otherUser.Health == "healthy" {
			msg := b.GetLocalString(lang, "healthy", map[string]string{"USERNAME": otherUsername})
			b.SendMessage(userMsg.Channel, msg)
			return true
		}

		if otherUser.Health == "dead" {
			b.UnbanUser(userMsg.Channel, otherUsername)
		}

		b.UpdateState(userMsg.Channel, func(chanState *db.ChannelState) {
			chanState.UserMap[otherUsername].Health = "healthy"
			chanState.UserMap[userMsg.Username].Stats["heals"]++
		})

		b.StopTimer(userMsg.Channel, otherUsername)

		msg := b.GetLocalString(lang, "on_heal", map[string]string{"USERNAME": otherUsername})
		b.SendMessage(userMsg.Channel, msg)

		return true

	case strings.HasPrefix(userMsg.Text, "!mend"):
		if user.Health != "deep_wound" {
			msg := b.GetLocalString(lang, "not_deep_wound", map[string]string{"USERNAME": userMsg.Username})
			b.SendMessage(userMsg.Channel, msg)

			return true
		}

		b.UpdateState(userMsg.Channel, func(chanState *db.ChannelState) {
			chanState.UserMap[userMsg.Username].Health = "injured"
		})

		b.StopTimer(userMsg.Channel, userMsg.Username)

		msg := b.GetLocalString(lang, "on_mend", map[string]string{"USERNAME": userMsg.Username})
		b.SendMessage(userMsg.Channel, msg)

		return true

	case strings.Contains(userMsg.Text, util.BotUsername) ||
		strings.Contains(userMsg.Text, "легион") ||
		strings.Contains(userMsg.Text, "лиджн") ||
		strings.Contains(userMsg.Text, "legion"):
		responseText, err := b.GenericResponse(lang, userMsg.Text)
		if err != nil {
			slog.Error("Failed to generate a generic response",
				slog.String("user", userMsg.Username),
				slog.String("text", userMsg.Text),
				slog.Any("error", err),
			)
			return true
		}

		b.SendMessage(userMsg.Channel, "@"+userMsg.Username+" "+responseText)

		return true
	}

	return false
}
