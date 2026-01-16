package alerts

import (
	"fmt"
	"log"
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

func (ta *TelegramAlerter) SendAlert(alertName, description string, severity int) error {
	hash := ta.createAlertHash(alertName, description)
	currentTime := time.Now()

	if state, exists := ta.activeAlerts[alertName]; exists {

		if state.Hash == hash && state.IsActive {

			if currentTime.Sub(state.LastUpdateTs) < ta.cooldownPeriod {

				state.LastUpdateTs = currentTime
				state.RepeatCount++
				ta.activeAlerts[alertName] = state
				return nil
			}

			state.RepeatCount++
			state.LastUpdateTs = currentTime
			ta.activeAlerts[alertName] = state

			updateMsg := fmt.Sprintf("🔁 [%s] %s (Active %s, repeats #%d)",
				severity, description,
				formatDuration(currentTime.Sub(state.FirstSeen)),
				state.RepeatCount)
			return ta.sendTelegramMessage(updateMsg)
		}

		if state.IsActive {
			ta.sendResolution(alertName, "changed",
				fmt.Sprintf("Issue changed: %s -> %s", state.Description, description))
		}
	}

	state := NewAlertState(alertName, Severity(severity), Description(description), nil, false, nil)

	ta.activeAlerts[alertName] = *state

	alertMsg := fmt.Sprintf("🚨 [%s] %s", severity, description)
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

	log.Printf("Alert %s resolved after %v", alertID, duration)
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

func (ta *TelegramAlerter) createAlertHash(alertName, description string) string {
	// normalized := normalizeMessage(message)
	hashInput := fmt.Sprintf("%s::%s", alertName, description)
	return fmt.Sprintf("%d", len(hashInput))
}

func (ta *TelegramAlerter) sendTelegramMessage(message string) error {
	msg := tgbotapi.NewMessage(ta.chatID, message)
	_, err := ta.bot.Send(msg)
	if err != nil {
		log.Printf("Failed to send Telegram message: %v", err)
		return err
	}
	log.Printf("Sent Telegram message: %s", message)
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
