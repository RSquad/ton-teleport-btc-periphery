package alerts

import (
	"fmt"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
)

type AlertDispatcherTg struct {
	mu              sync.RWMutex
	telegramAlerter *TelegramAlerter
	staleDuration   time.Duration
	cleanupInterval time.Duration
	stopChan        chan struct{}
}

func NewAlertDispatcherTg(telegramAlerter *TelegramAlerter) *AlertDispatcherTg {
	if telegramAlerter == nil {
		logger.Log.Panic().
			Str("component", "AlertDispatcherTg").
			Msg("telegramAlerter cannot be nil")
	}

	dispatcher := &AlertDispatcherTg{
		telegramAlerter: telegramAlerter,
		staleDuration:   3 * time.Hour, // TODO: make configurable
		cleanupInterval: 1 * time.Hour, // TODO: make configurable
		stopChan:        make(chan struct{}),
	}

	// Start background cleanup goroutine
	go dispatcher.runBackgroundTasks()

	return dispatcher
}

func NewAlertDispatcherTgWithConfig(
	telegramAlerter *TelegramAlerter,
	staleDuration,
	cleanupInterval time.Duration,
) *AlertDispatcherTg {
	if telegramAlerter == nil {
		logger.Log.Error().
			Str("component", "AlertDispatcherTg").
			Msg("telegramAlerter cannot be nil")
	}

	dispatcher := &AlertDispatcherTg{
		telegramAlerter: telegramAlerter,
		staleDuration:   staleDuration,
		cleanupInterval: cleanupInterval,
		stopChan:        make(chan struct{}),
	}

	go dispatcher.runBackgroundTasks()

	return dispatcher
}

func (d *AlertDispatcherTg) OnAlert(state *AlertState) error {
	if state == nil {
		return nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if d.telegramAlerter == nil {
		return nil
	}

	return d.telegramAlerter.SendAlert(state)
}

func (d *AlertDispatcherTg) ResolveAlert(alertID, resolutionDetails string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.telegramAlerter == nil {
		return nil
	}

	return d.telegramAlerter.ResolveAlert(alertID, resolutionDetails)
}

func (d *AlertDispatcherTg) GetActiveAlerts() []AlertState {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.telegramAlerter == nil {
		return []AlertState{}
	}

	return d.telegramAlerter.GetActiveAlerts()
}

func (d *AlertDispatcherTg) SetTelegramAlerter(alerter *TelegramAlerter) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.telegramAlerter = alerter
}

func (d *AlertDispatcherTg) SetStaleDuration(duration time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.staleDuration = duration
}

func (d *AlertDispatcherTg) runBackgroundTasks() {
	cleanupTicker := time.NewTicker(d.cleanupInterval)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-cleanupTicker.C:
			d.runCleanup()
		case <-d.stopChan:
			return
		}
	}
}

func (d *AlertDispatcherTg) runCleanup() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.telegramAlerter != nil {
		d.telegramAlerter.AutoResolveStaleAlerts(d.staleDuration)
	}
}

func (d *AlertDispatcherTg) Stop() {
	close(d.stopChan)

	d.mu.Lock()
	defer d.mu.Unlock()

	if d.telegramAlerter != nil {
		activeAlerts := d.telegramAlerter.GetActiveAlerts()
		if len(activeAlerts) > 0 {
			summary := "📋 Alert Dispatcher Stopping\n"
			summary += fmt.Sprintf("Active alerts: %d\n", len(activeAlerts))
			for _, alert := range activeAlerts {
				summary += fmt.Sprintf("• %s: %s (active for %s)\n",
					alert.Name,
					alert.Description,
					formatDuration(time.Since(alert.FirstSeen)))
			}
			d.telegramAlerter.sendTelegramMessage(summary)
		}
	}
}
