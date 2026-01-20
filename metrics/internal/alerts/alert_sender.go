package alerts

import (
	"fmt"
	"regexp"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
)

type TelegramAlerter struct {
	bot            *tgbotapi.BotAPI
	chatID         int64
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

func NewTelegramAlerter(token string, chatID int64, cooldown time.Duration) (*TelegramAlerter, error) {
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
	hash := ta.createAlertHash(state.Name, state.Description)
	currentTime := time.Now()

	if state, exists := ta.activeAlerts[state.Name]; exists {

		if ta.activeAlerts[state.Name].Hash == hash && state.IsActive { // TODO: check comparison

			if currentTime.Sub(state.LastUpdateTs) < ta.cooldownPeriod {

				state.LastUpdateTs = currentTime
				state.RepeatCount++
				ta.activeAlerts[state.Name] = state
				return nil
			}

			state.RepeatCount++
			state.LastUpdateTs = currentTime
			ta.activeAlerts[state.Name] = state

			updateMsg := fmt.Sprintf("🔁 [%s] %s (Active %s, repeats #%d)",
				state.Severity, state.Description,
				formatDuration(currentTime.Sub(state.FirstSeen)),
				state.RepeatCount)
			return ta.sendTelegramMessage(updateMsg)
		}

		if state.IsActive {
			ta.sendResolution(state.Name, "changed",
				fmt.Sprintf("Issue changed: %s -> %s", ta.activeAlerts[state.Name].Description, state.Description))
		}
	}

	ta.activeAlerts[state.Name] = *state

	alertMsg := fmt.Sprintf("🚨 [%s] %s", state.Severity, state.Description)
	return ta.sendTelegramMessage(alertMsg)
}

func (ta *TelegramAlerter) sendResolution(alertID, resolution, details string) error {
	state, exists := ta.activeAlerts[alertID]
	if !exists || !state.IsActive {
		return fmt.Errorf("alert %s is not active", alertID)
	}

	duration := time.Since(state.FirstSeen)
	resolutionMsg := fmt.Sprintf("✅ [RESOLVED: %s] %s\n📊 Duration: %s\n📝 Details: %s",
		resolution, state.Description, formatDuration(duration), details)

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
	// normalized := normalizeMessage(message)
	hashInput := fmt.Sprintf("%s::%s", alertName, description)
	return fmt.Sprintf("%d", len(hashInput))
}

func (ta *TelegramAlerter) sendTelegramMessage(message string) error {
	msg := tgbotapi.NewMessage(ta.chatID, message)
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
	ticker := time.NewTicker(1 * time.Hour)

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
