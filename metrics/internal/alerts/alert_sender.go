package alerts

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
)

const timeLayout = "Mon 2 Jan 15:04:05 MST 2006"

type TelegramAlerter struct {
	bot            *tgbotapi.BotAPI
	chatID         int
	activeAlerts   map[string]AlertState
	cooldownPeriod time.Duration
}

func removeTimestamps(text string) string {
	re := regexp.MustCompile(`\d{1,2}:\d{2}(:\d{2})?`)
	return re.ReplaceAllString(text, "")
}

func removeNumbers(text string) string {
	re := regexp.MustCompile(`\d+`)
	return re.ReplaceAllString(text, "")
}

func NewTelegramAlerter(token string, chatID int, cooldown time.Duration) (*TelegramAlerter, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	logger.Log.Info().
		Str("component", "TelegramAlerter").
		Msg(fmt.Sprintf("Authorized on account %s", bot.Self.UserName))

	return &TelegramAlerter{
		bot:            bot,
		chatID:         chatID,
		activeAlerts:   make(map[string]AlertState),
		cooldownPeriod: cooldown,
	}, nil
}

func (ta *TelegramAlerter) SendAlert(state *AlertState) error {
	if state.Severity == 1 { // skipping INFO alerts
		return nil
	}

	hash := ta.createAlertHash(state.Name, state.Description)
	currentTime := time.Now()

	if existingState, exists := ta.activeAlerts[state.Name]; exists {
		if existingState.Hash == hash {
			if existingState.IsActive {
				if currentTime.Sub(existingState.LastUpdateTs) < ta.cooldownPeriod {
					existingState.LastUpdateTs = currentTime
					existingState.RepeatCount++
					ta.activeAlerts[state.Name] = existingState
					return nil
				}

				existingState.RepeatCount++
				existingState.LastUpdateTs = currentTime
				ta.activeAlerts[state.Name] = existingState

				updateMsg := ta.formatUpdateMessage(&existingState)
				return ta.sendTelegramMessage(updateMsg)
			} else {
				state.FirstSeen = currentTime
				state.LastUpdateTs = currentTime
				state.IsActive = true
				state.Hash = hash
				state.RepeatCount = 1
				ta.activeAlerts[state.Name] = *state

				alertMsg := ta.formatAlertMessage(state)
				return ta.sendTelegramMessage(alertMsg)
			}
		} else {
			if existingState.IsActive {
				ta.sendResolution(state.Name, "changed",
					fmt.Sprintf("Issue changed: %s -> %s", existingState.Description, state.Description))
			}

			state.FirstSeen = currentTime
			state.LastUpdateTs = currentTime
			state.IsActive = true
			state.Hash = hash
			state.RepeatCount = 1
			ta.activeAlerts[state.Name] = *state

			alertMsg := ta.formatAlertMessage(state)
			return ta.sendTelegramMessage(alertMsg)
		}
	}

	state.FirstSeen = currentTime
	state.LastUpdateTs = currentTime
	state.IsActive = true
	state.Hash = hash
	state.RepeatCount = 1
	ta.activeAlerts[state.Name] = *state

	alertMsg := ta.formatAlertMessage(state)
	return ta.sendTelegramMessage(alertMsg)
}

func (ta *TelegramAlerter) formatAlertMessage(state *AlertState) string {
	currentTime := time.Now().UTC()

	icon := ta.choseIcon(state.Severity)

	name := escapeMarkdownV2(state.Name)
	date := escapeMarkdownV2(currentTime.Format(timeLayout))
	description := escapeMarkdownV2(fmt.Sprintf("%v", state.Description))

	return fmt.Sprintf(`%s *FIRING*
*%s*
%s
*Date:* %s
*Description:* %s`,
		icon,
		name,
		"\\-\\-\\-\\-\\-",
		date,
		description,
	)
}

func (ta *TelegramAlerter) choseIcon(severity Severity) string {
	var icon string
	switch severity {
	case SEVERITY_CRITICAL:
		icon = "🔴"
	case SEVERITY_WARNING:
		icon = "⚠️"
	default:
		icon = "ℹ️"
	}

	return icon
}

func (ta *TelegramAlerter) formatUpdateMessage(state *AlertState) string {
	currentTime := time.Now().UTC()
	duration := currentTime.Sub(state.FirstSeen)

	icon := ta.choseIcon(state.Severity)

	name := escapeMarkdownV2(state.Name)
	date := escapeMarkdownV2(currentTime.Format(timeLayout))
	durationStr := escapeMarkdownV2(formatDuration(duration))
	description := escapeMarkdownV2(fmt.Sprintf("%v", state.Description))

	return fmt.Sprintf(`%s *STILL FIRING* \\(#%d\\)
*%s*
%s
*Date:* %s
*Duration:* %s
*Description:* %s`,
		icon,
		state.RepeatCount,
		name,
		"\\-\\-\\-\\-\\-",
		date,
		durationStr,
		description,
	)
}

func (ta *TelegramAlerter) sendResolution(alertID, resolution, details string) error {
	state, exists := ta.activeAlerts[alertID]
	if !exists || !state.IsActive {
		return fmt.Errorf("alert %s is not active", alertID)
	}

	duration := time.Since(state.FirstSeen)
	currentTime := time.Now().UTC()

	icon := ta.choseIcon(state.Severity)

	name := escapeMarkdownV2(state.Name)
	date := escapeMarkdownV2(currentTime.Format(timeLayout))
	resolutionEscaped := escapeMarkdownV2(resolution)
	durationStr := escapeMarkdownV2(formatDuration(duration))
	detailsEscaped := escapeMarkdownV2(details)

	resolutionMsg := fmt.Sprintf(`✅ *%s RESOLVED*
*%s*
%s
*Date:* %s
*Resolution:* %s
*Duration:* %s
*Details:* %s`,
		icon,
		name,
		"\\-\\-\\-\\-\\-",
		date,
		resolutionEscaped,
		durationStr,
		detailsEscaped,
	)

	err := ta.sendTelegramMessage(resolutionMsg)
	if err != nil {
		return err
	}

	state.IsActive = false
	ta.activeAlerts[alertID] = state

	logger.Log.Info().
		Str("component", "TelegramAlerter").
		Str("alertID", alertID).
		Str("resolution", resolution).
		Str("details", details).
		Msg("Alert resolved")
	return nil
}

func escapeMarkdownV2(text string) string {
	result := strings.Builder{}

	for _, r := range text {
		char := string(r)
		switch char {
		case "_", "*", "[", "]", "(", ")", "~", "`", ">",
			"#", "+", "-", "=", "|", "{", "}", ".", "!":
			result.WriteString("\\" + char)
		case "\\":
			result.WriteString("\\\\")
		default:
			result.WriteString(char)
		}
	}

	return result.String()
}

func (ta *TelegramAlerter) ResolveAlert(alertID, resolutionDetails string) error {
	return ta.sendResolution(alertID, "resolved", resolutionDetails)
}

func (ta *TelegramAlerter) AutoResolveStaleAlerts(staleDuration time.Duration) {
	cutoff := time.Now().Add(-staleDuration)

	for alertID, state := range ta.activeAlerts {
		if state.IsActive && state.LastUpdateTs.Before(cutoff) {
			ta.sendResolution(alertID, "autoresolved",
				fmt.Sprintf("Inactive for %v", staleDuration))
		}
	}
}

func (ta *TelegramAlerter) GetActiveAlerts() []AlertState {
	result := make([]AlertState, 0, len(ta.activeAlerts))
	for _, state := range ta.activeAlerts {
		if state.IsActive {
			result = append(result, state)
		}
	}
	return result
}

func (ta *TelegramAlerter) createAlertHash(alertName string, description Description) string {
	normalized := normalizeMessage(description)
	return fmt.Sprintf("%s::%s", alertName, normalized)
}

func normalizeMessage(d Description) string {
	s := fmt.Sprintf("%v", d)
	s = removeTimestamps(s)
	s = removeNumbers(s)
	return s
}

func (ta *TelegramAlerter) sendTelegramMessage(message string) error {
	msg := tgbotapi.NewMessage(int64(ta.chatID), message)
	msg.ParseMode = "MarkdownV2"
	msg.DisableWebPagePreview = false
	_, err := ta.bot.Send(msg)
	if err != nil {
		logger.Log.Error().Str("component", "TelegramAlerter").Err(err).Msg("Failed to send Telegram message")
		return err
	}
	logger.Log.Info().Str("component", "TelegramAlerter").Msg("Message sent")
	return nil
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0f сек", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.0f мин", d.Minutes())
	} else {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		return fmt.Sprintf("%d ч %d мин", hours, minutes)
	}
}

func (ta *TelegramAlerter) RunCleanup() {
	ticker := time.NewTicker(1 * time.Hour) // TODO: make configurable

	for range ticker.C {
		ta.cleanupOldAlerts()
		ta.AutoResolveStaleAlerts(ta.cooldownPeriod * time.Hour) // Autoresolve alerts inactive for N hours
	}
}

func (ta *TelegramAlerter) cleanupOldAlerts() {
	cutoff := time.Now().Add(-7 * 24 * time.Hour) // TODO: make configurable

	for alertID, state := range ta.activeAlerts {
		if !state.IsActive && state.LastUpdateTs.Before(cutoff) {
			delete(ta.activeAlerts, alertID)
		}
	}
}
