package alerts

import (
	"fmt"
	"regexp"
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
		if existingState.Hash == hash && existingState.IsActive {
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
			fmt.Println("MESSAGE: ", updateMsg)
			return ta.sendTelegramMessage(updateMsg)
		}

		if existingState.IsActive {
			ta.sendResolution(state.Name, "changed",
				fmt.Sprintf("Issue changed: %s -> %s", existingState.Description, state.Description))
		}
	}

	state.FirstSeen = currentTime
	state.LastUpdateTs = currentTime
	state.IsActive = true
	ta.activeAlerts[state.Name] = *state

	alertMsg := ta.formatAlertMessage(state)
	return ta.sendTelegramMessage(alertMsg)
}

func (ta *TelegramAlerter) formatAlertMessage(state *AlertState) string {
	currentTime := time.Now().UTC()

	var icon string
	switch state.Severity {
	case 2:
		icon = "🔴"
	case 3:
		icon = "⚠️"
	default:
		icon = "ℹ️"
	}

	return fmt.Sprintf(`%s FIRING
%s
-----
Date: %s
Description: %s`,
		icon,
		state.Name,
		currentTime.Format(timeLayout),
		state.Description,
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

	return fmt.Sprintf(`%s STILL FIRING (#%d)
%s
-----
Date: %s
Duration: %s
Description: %s`,
		icon,
		state.RepeatCount+1,
		state.Name,
		currentTime.Format(timeLayout),
		formatDuration(duration),
		state.Description,
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

	resolutionMsg := fmt.Sprintf(`✅ %s RESOLVED
%s
-----
Date: %s
Resolution: %s
Duration: %s
Details: %s`,
		icon,
		state.Name,
		currentTime.Format(timeLayout),
		resolution,
		formatDuration(duration),
		details,
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

func (ta *TelegramAlerter) resolveAlert(alertID, resolutionDetails string) error {
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
	// normalized := normalizeMessage(message)
	hashInput := fmt.Sprintf("%s::%s", alertName, description)
	return fmt.Sprintf("%d", len(hashInput))
}

func (ta *TelegramAlerter) sendTelegramMessage(message string) error {
	msg := tgbotapi.NewMessage(int64(ta.chatID), message)
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
