package watchdog

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
)

type WatchdogOverdueCallback func(id string, overdue time.Duration)

type Element struct {
	lastHeartbeat time.Time
	lastFire      time.Time
	period        time.Duration
}

type Watchdog struct {
	scanPeriod      time.Duration
	firePeriod      time.Duration
	overdueCallback WatchdogOverdueCallback

	mu            sync.Mutex
	lastHeartbeat map[string]Element
	wg            sync.WaitGroup
}

func NewWatchdog(
	scanPeriod time.Duration,
	firePeriod time.Duration,
	overdueCallback WatchdogOverdueCallback,
) (*Watchdog, error) {
	if scanPeriod <= 0 {
		return nil, errors.New("watchdog scanPeriod must be > 0")
	}

	if firePeriod <= scanPeriod {
		return nil, fmt.Errorf("watchdog firePeriod must be > %d", scanPeriod)
	}

	if overdueCallback == nil {
		return nil, errors.New("watchdog overdue callback is null")
	}

	return &Watchdog{
		scanPeriod:      scanPeriod,
		firePeriod:      firePeriod,
		overdueCallback: overdueCallback,
		lastHeartbeat:   make(map[string]Element),
	}, nil
}

var globalInstance *Watchdog = nil

func InitGlobal(
	scanPeriod time.Duration,
	firePeriod time.Duration,
	overdueCallback WatchdogOverdueCallback,
) error {
	var err error
	globalInstance, err = NewWatchdog(scanPeriod, firePeriod, overdueCallback)

	if err != nil {
		logger.Log.Error().Str("component", "WATCHDOG").Msgf("Global initialization error: %s", err)
		return err
	}

	logger.Log.Info().Str("component", "WATCHDOG").Msg("Global initialization completed successfully.")

	return nil
}

func Global() *Watchdog {
	return globalInstance
}

func (w *Watchdog) Start(ctx context.Context) {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()

		logger.Log.Info().Str("component", "WATCHDOG").Msg("Started")

		t := time.NewTicker(w.scanPeriod / 2)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				w.scan()
			}
		}
	}()
}

func (w *Watchdog) Watch(id string, period time.Duration) {
	logger.Log.Debug().Str("component", "WATCHDOG").Msgf("Watch '%s'", id)

	now := time.Now()

	element := Element{
		lastHeartbeat: now,
		lastFire:      now,
		period:        period,
	}

	w.mu.Lock()
	w.lastHeartbeat[id] = element
	w.mu.Unlock()
}

func (w *Watchdog) Unwatch(id string) {
	logger.Log.Debug().Str("component", "WATCHDOG").Msgf("Unwatch '%s'", id)

	w.mu.Lock()
	delete(w.lastHeartbeat, id)
	w.mu.Unlock()
}

func (w *Watchdog) Heartbeat(id string) {
	logger.Log.Debug().Str("component", "WATCHDOG").Msgf("Heartbeat '%s'", id)

	now := time.Now()
	w.mu.Lock()
	if e, ok := w.lastHeartbeat[id]; ok {
		e.lastHeartbeat = now
		w.lastHeartbeat[id] = e
	}
	w.mu.Unlock()
}

func (w *Watchdog) scan() {
	now := time.Now()
	var overdueIDs []string
	var overdues []time.Duration

	{
		w.mu.Lock()
		defer w.mu.Unlock()

		for id, last := range w.lastHeartbeat {

			// Check overdue
			if d := now.Sub(last.lastHeartbeat); d > last.period {
				// Check firePeriod
				if fd := now.Sub(last.lastFire); fd > w.firePeriod {
					overdueIDs = append(overdueIDs, id)
					overdues = append(overdues, d)
				}
			}
		}
	}

	// Fire callbacks
	for i, id := range overdueIDs {
		od := overdues[i]
		w.overdueCallback(id, od)

		e := w.lastHeartbeat[id]
		e.lastFire = now
		w.lastHeartbeat[id] = e
	}
}
