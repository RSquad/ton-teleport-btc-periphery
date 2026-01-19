package alerts

import (
	"sync"
)

type AlertDispatcherTg struct {
	mu              sync.RWMutex
	telegramAlerter *TelegramAlerter
}

func NewAlertDispatcherTg() *AlertDispatcherTg {
	return &AlertDispatcherTg{
		telegramAlerter: nil,
	}
}

func (d *AlertDispatcherTg) OnAlert(state *AlertState) error {
	if state == nil {
		return nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.telegramAlerter.SendAlert(state)

	return nil
}
