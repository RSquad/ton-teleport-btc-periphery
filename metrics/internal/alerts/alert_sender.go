package alerts

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
)

const timeLayout = "Mon 2 Jan 15:04:05 MST 2006"

type TelegramAlerter struct {
	bot                 *tgbotapi.BotAPI
	chatID              int
	activeAlerts        map[string]AlertState
	cooldownPeriod      time.Duration
	inactiveAlertPeriod time.Duration
}

func removeTimestamps(text string) string {
	re := regexp.MustCompile(`\d{1,2}:\d{2}(:\d{2})?`)
	return re.ReplaceAllString(text, "")
}

func removeNumbers(text string) string {
	re := regexp.MustCompile(`\d+`)
	return re.ReplaceAllString(text, "")
}

func NewTelegramAlerter(token string, chatID int, cooldown time.Duration, inactiveAlertPeriod time.Duration) (*TelegramAlerter, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	logger.Log.Info().
		Str("component", "TelegramAlerter").
		Msg(fmt.Sprintf("Authorized on account %s", bot.Self.UserName))

	return &TelegramAlerter{
		bot:                 bot,
		chatID:              chatID,
		activeAlerts:        make(map[string]AlertState),
		cooldownPeriod:      cooldown,
		inactiveAlertPeriod: inactiveAlertPeriod,
	}, nil
}

func (ta *TelegramAlerter) SendAlert(state *AlertState) error {
	if state.Severity == SEVERITY_OK {
		if existingState, exists := ta.activeAlerts[state.Name]; exists {
			if existingState.IsActive && existingState.Severity != SEVERITY_OK {
				return ta.sendResolution(state.Name, "resolved",
					fmt.Sprintf("Issue resolved: %s", state.Description))
			}
		}
		return nil
	}

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
				if existingState.Severity != SEVERITY_OK {
					ta.sendResolution(state.Name, "changed",
						fmt.Sprintf("Issue changed: %s -> %s", existingState.Description, state.Description))
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

	name := html.EscapeString(state.Name)
	date := html.EscapeString(currentTime.Format(timeLayout))
	description := ensureHTMLLinks(fmt.Sprintf("%v", state.Description))

	return fmt.Sprintf(`%s <b>FIRING</b>
<b>%s</b>
-----
<b>Date:</b> %s
<b>Description:</b> %s`,
		icon,
		name,
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
	case SEVERITY_OK:
		icon = "✅"
	default:
		icon = "ℹ️"
	}

	return icon
}

func (ta *TelegramAlerter) formatUpdateMessage(state *AlertState) string {
	currentTime := time.Now().UTC()
	duration := currentTime.Sub(state.FirstSeen)

	icon := ta.choseIcon(state.Severity)

	name := html.EscapeString(state.Name)
	date := html.EscapeString(currentTime.Format(timeLayout))
	durationStr := html.EscapeString(formatDuration(duration))
	description := ensureHTMLLinks(fmt.Sprintf("%v", state.Description))

	return fmt.Sprintf(`%s <b>STILL FIRING</b> (#%d)
<b>%s</b>
-----
<b>Date:</b> %s
<b>Duration:</b> %s
<b>Description:</b> %s`,
		icon,
		state.RepeatCount,
		name,
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

	icon := "✅"

	name := html.EscapeString(state.Name)
	date := html.EscapeString(currentTime.Format(timeLayout))
	resolutionEscaped := html.EscapeString(resolution)
	durationStr := html.EscapeString(formatDuration(duration))
	detailsEscaped := ensureHTMLLinks(details)

	resolutionMsg := fmt.Sprintf(`%s <b>RESOLVED</b>
<b>%s</b>
-----
<b>Date:</b> %s
<b>Resolution:</b> %s
<b>Duration:</b> %s
<b>Details:</b> %s`,
		icon,
		name,
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
	state.LastUpdateTs = currentTime
	ta.activeAlerts[alertID] = state

	logger.Log.Info().
		Str("component", "TelegramAlerter").
		Str("alertID", alertID).
		Str("resolution", resolution).
		Str("details", details).
		Msg("Alert resolved")
	return nil
}

func ensureHTMLLinks(text string) string {
	if strings.Contains(text, "<a ") || strings.Contains(text, "href=") {
		return text
	}

	text = html.EscapeString(text)

	urlRegex := regexp.MustCompile(`(https?://[^\s<]+)`)

	result := urlRegex.ReplaceAllStringFunc(text, func(url string) string {
		cleanURL := strings.TrimRight(url, ".,;!?)")
		return fmt.Sprintf(`<a href="%s">%s</a>`, cleanURL, cleanURL)
	})

	return result
}

func (ta *TelegramAlerter) ResolveAlert(alertID string, description string) error {
	if state, exists := ta.activeAlerts[alertID]; exists && state.IsActive {
		resolvedState := &AlertState{
			Name:        alertID,
			Description: Description(fmt.Sprintf("Resolved: %s", description)),
			Severity:    SEVERITY_OK,
		}
		return ta.SendAlert(resolvedState)
	}
	return nil
}

func (ta *TelegramAlerter) AutoResolveStaleAlerts(staleDuration time.Duration) {
	cutoff := time.Now().Add(-staleDuration)

	for alertID, state := range ta.activeAlerts {
		if state.IsActive && state.LastUpdateTs.Before(cutoff) {
			if state.Severity != SEVERITY_OK {
				resolvedState := &AlertState{
					Name:        alertID,
					Description: Description(fmt.Sprintf("Inactive for %v", staleDuration)),
					Severity:    SEVERITY_OK,
				}
				ta.SendAlert(resolvedState)
			}
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
	msg.ParseMode = "HTML" // Используем HTML разметку
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
	ticker := time.NewTicker(10 * time.Second)

	for range ticker.C {
		ta.cleanupOldAlerts()
		if ta.inactiveAlertPeriod > 0 {
			ta.AutoResolveStaleAlerts(ta.inactiveAlertPeriod * time.Hour)
		}
	}
}

func (ta *TelegramAlerter) cleanupOldAlerts() {
	cutoff := time.Now().Add(-7 * ta.inactiveAlertPeriod * time.Hour)

	for alertID, state := range ta.activeAlerts {
		if !state.IsActive && state.LastUpdateTs.Before(cutoff) {
			delete(ta.activeAlerts, alertID)
		}
	}
}
