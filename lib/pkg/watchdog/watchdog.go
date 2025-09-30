// Package watchdog provides a lightweight heartbeat watchdog for tracking
// multiple targets. Each target is expected to send periodic heartbeats; if a
// target exceeds its allowed period, a user-supplied callback is invoked.

package watchdog

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
)

// OverdueCallback is called when a watched target becomes overdue.
// The id is the target identifier and overdue is how long the target
// has exceeded its expected heartbeat period.
type OverdueCallback func(id string, overdue time.Duration)

// WatchedTarget describes a single watched target's timing data.
// It is stored internally by Watchdog.
type WatchedTarget struct {
	lastHeartbeat time.Time
	lastFire      time.Time
	period        time.Duration
}

// Watchdog monitors many targets and invokes a callback when any target
// becomes overdue. Watchdog is safe for concurrent use by multiple goroutines.
//
// Scanning:
//   - The Watchdog scans all targets every scanPeriod.
//   - For each overdue target, overdueCallback is invoked at most once per
//     firePeriod (per target).
//
// Self heartbeat:
//   - The Watchdog emits a log "SELF HEARTBEAT (is alive)" no more frequently
//     than selfHeartbeatPeriod to indicate the Watchdog is alive.
type Watchdog struct {
	scanPeriod          time.Duration
	firePeriod          time.Duration
	selfHeartbeatPeriod time.Duration
	selfLastHeartbeat   time.Time
	overdueCallback     OverdueCallback

	mu      sync.Mutex
	targets map[string]WatchedTarget
	wg      sync.WaitGroup
}

// New constructs a Watchdog that scans targets every scanPeriod and,
// for each target that misses its heartbeat period, invokes overdueCallback,
// throttled so it cannot fire more frequently than firePeriod per target.
//
// Preconditions:
//   - scanPeriod must be > 0;
//   - firePeriod must be > scanPeriod;
//   - overdueCallback must be non-nil.
//
// selfHeartbeatPeriod controls how often the Watchdog logs its own liveness
// (zero or negative disables it). Returns an error if validation fails.
func new(
	scanPeriod time.Duration,
	firePeriod time.Duration,
	selfHeartbeatPeriod time.Duration,
	overdueCallback OverdueCallback,
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
		scanPeriod:          scanPeriod,
		firePeriod:          firePeriod,
		selfHeartbeatPeriod: selfHeartbeatPeriod,
		selfLastHeartbeat:   time.Now(),
		overdueCallback:     overdueCallback,
		targets:             make(map[string]WatchedTarget),
	}, nil
}

var globalInstanceMU sync.Mutex
var globalInstance *Watchdog = nil

// InitGlobalAndStart initializes the package-global Watchdog singleton and
// starts its background scan loop. If initialization has already occurred,
// it returns an error. The scan loop runs until ctx is canceled.
//
// Typical usage:
//
//	err := watchdog.InitGlobalAndStart(… , ctx)
//	if err != nil {...}
//	...
//
//	// In every watched goroutine
//	watchdog.Global().Watch("my_goroutine_id", 5*time.Second)
//	...
//	watchdog.Global().Heartbeat("my_goroutine_id")
//	...
func Init(
	scanPeriod time.Duration,
	firePeriod time.Duration,
	selfHeartbeatPeriod time.Duration,
	overdueCallback OverdueCallback,
	ctx context.Context,
) error {
	globalInstanceMU.Lock()
	defer globalInstanceMU.Unlock()

	var err error

	if globalInstance != nil {
		return errors.New("only a single Watchdog instance is supported per application")
	}

	globalInstance, err = new(
		scanPeriod,
		firePeriod,
		selfHeartbeatPeriod,
		overdueCallback,
	)

	if err != nil {
		return err
	}

	globalInstance.start(ctx)

	return nil
}

// Global returns the package-global Watchdog, or nil if it has not been
// initialized via InitGlobalAndStart.
func Global() *Watchdog {
	globalInstanceMU.Lock()
	defer globalInstanceMU.Unlock()

	if globalInstance == nil {
		logger.Log.Error().Str("component", "WATCHDOG").Msg("Global Watchdog is null. Please call watchdog.InitGlobalAndStart")
	}

	return globalInstance
}

// start launches the Watchdog's background scan loop in a goroutine.
// It stops when ctx is canceled. Intended to be called by InitGlobalAndStart.
func (w *Watchdog) start(ctx context.Context) {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()

		logger.Log.Info().Str("component", "WATCHDOG").Msg("Started")

		t := time.NewTicker(w.scanPeriod)
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

	target := WatchedTarget{
		lastHeartbeat: now,
		lastFire:      now,
		period:        period,
	}

	w.mu.Lock()
	w.targets[id] = target
	w.mu.Unlock()
}

func (w *Watchdog) Unwatch(id string) {
	logger.Log.Debug().Str("component", "WATCHDOG").Msgf("Unwatch '%s'", id)

	w.mu.Lock()
	delete(w.targets, id)
	w.mu.Unlock()
}

func (w *Watchdog) Heartbeat(id string) {
	logger.Log.Debug().Str("component", "WATCHDOG").Msgf("Heartbeat '%s'", id)

	now := time.Now()
	w.mu.Lock()
	if e, ok := w.targets[id]; ok {
		e.lastHeartbeat = now
		w.targets[id] = e
	}
	w.mu.Unlock()
}

func (w *Watchdog) scan() {
	now := time.Now()
	var overdueIDs []string
	var overdues []time.Duration

	w.mu.Lock()
	defer w.mu.Unlock()

	for id, target := range w.targets {
		// Check overdue
		if dt := time.Since(target.lastHeartbeat); dt > target.period {
			// Check firePeriod
			if fd := now.Sub(target.lastFire); fd > w.firePeriod {
				overdueIDs = append(overdueIDs, id)
				overdues = append(overdues, dt)
			}
		}
	}

	// Fire callbacks
	for i, id := range overdueIDs {
		od := overdues[i]
		w.overdueCallback(id, od)

		t := w.targets[id]
		t.lastFire = now
		w.targets[id] = t
	}

	// Self heartbeat
	if time.Since(w.selfLastHeartbeat) > w.selfHeartbeatPeriod {
		w.selfLastHeartbeat = now
		logger.Log.Info().Str("component", "WATCHDOG").Msgf("Watchdog is alive")
	}
}
