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
// It is stored internally by Manager.
type WatchedTarget struct {
	lastHeartbeat time.Time
	lastFire      time.Time
	period        time.Duration
}

// Manager monitors many targets and invokes a callback when any target
// becomes overdue. Manager is safe for concurrent use by multiple goroutines.
//
// Scanning:
//   - The manager scans all targets every scanPeriod.
//   - For each overdue target, overdueCallback is invoked at most once per
//     firePeriod (per target).
//
// Self heartbeat:
//   - The manager emits a log "SELF HEARTBEAT (is alive)" no more frequently
//     than selfHeartbeatPeriod to indicate the Watchdog is alive.
type Manager struct {
	scanPeriod          time.Duration
	firePeriod          time.Duration
	selfHeartbeatPeriod time.Duration
	selfLastHeartbeat   time.Time
	overdueCallback     OverdueCallback

	mu      sync.Mutex
	targets map[string]WatchedTarget
	wg      sync.WaitGroup
}

// NewManager constructs a Watchdog Manager that scans targets every scanPeriod and,
// for each target that misses its heartbeat period, invokes overdueCallback,
// throttled so it cannot fire more frequently than firePeriod per target.
//
// Preconditions:
//   - scanPeriod must be > 0;
//   - firePeriod must be > scanPeriod;
//   - overdueCallback must be non-nil.
//
// selfHeartbeatPeriod controls how often the manager logs its own liveness
// (zero or negative disables it). Returns an error if validation fails.
func NewManager(
	scanPeriod time.Duration,
	firePeriod time.Duration,
	selfHeartbeatPeriod time.Duration,
	overdueCallback OverdueCallback,
) (*Manager, error) {
	if scanPeriod <= 0 {
		return nil, errors.New("watchdog scanPeriod must be > 0")
	}

	if firePeriod <= scanPeriod {
		return nil, fmt.Errorf("watchdog firePeriod must be > %d", scanPeriod)
	}

	if overdueCallback == nil {
		return nil, errors.New("watchdog overdue callback is null")
	}

	return &Manager{
		scanPeriod:          scanPeriod,
		firePeriod:          firePeriod,
		selfHeartbeatPeriod: selfHeartbeatPeriod,
		selfLastHeartbeat:   time.Now(),
		overdueCallback:     overdueCallback,
		targets:             make(map[string]WatchedTarget),
	}, nil
}

var globalInstanceMU sync.Mutex
var globalInstance *Manager = nil

// InitGlobalAndStart initializes the package-global Manager singleton and
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
func InitGlobalAndStart(
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
		return errors.New("watchdog.InitGlobal can be called only once")
	}

	globalInstance, err = NewManager(
		scanPeriod,
		firePeriod,
		selfHeartbeatPeriod,
		overdueCallback,
	)

	if err != nil {
		logger.Log.Error().Str("component", "WATCHDOG").Msgf("Global initialization error: %s", err)
		return err
	}

	logger.Log.Info().Str("component", "WATCHDOG").Msg("Global initialization completed successfully.")
	globalInstance.start(ctx)

	return nil
}

// Global returns the package-global Manager, or nil if it has not been
// initialized via InitGlobalAndStart.
func Global() *Manager {
	globalInstanceMU.Lock()
	defer globalInstanceMU.Unlock()

	if globalInstance == nil {
		logger.Log.Error().Str("component", "WATCHDOG").Msg("Global Manager is null. Please call watchdog.InitGlobalAndStart")
	}

	return globalInstance
}

// start launches the Manager's background scan loop in a goroutine.
// It stops when ctx is canceled. Intended to be called by InitGlobalAndStart.
func (manager *Manager) start(ctx context.Context) {
	manager.wg.Add(1)
	go func() {
		defer manager.wg.Done()

		logger.Log.Info().Str("component", "WATCHDOG").Msg("Started")

		t := time.NewTicker(manager.scanPeriod)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				manager.scan()
			}
		}
	}()
}

func (manager *Manager) Watch(id string, period time.Duration) {
	logger.Log.Debug().Str("component", "WATCHDOG").Msgf("Watch '%s'", id)

	now := time.Now()

	target := WatchedTarget{
		lastHeartbeat: now,
		lastFire:      now,
		period:        period,
	}

	manager.mu.Lock()
	manager.targets[id] = target
	manager.mu.Unlock()
}

func (manager *Manager) Unwatch(id string) {
	logger.Log.Debug().Str("component", "WATCHDOG").Msgf("Unwatch '%s'", id)

	manager.mu.Lock()
	delete(manager.targets, id)
	manager.mu.Unlock()
}

func (manager *Manager) Heartbeat(id string) {
	logger.Log.Debug().Str("component", "WATCHDOG").Msgf("Heartbeat '%s'", id)

	now := time.Now()
	manager.mu.Lock()
	if e, ok := manager.targets[id]; ok {
		e.lastHeartbeat = now
		manager.targets[id] = e
	}
	manager.mu.Unlock()
}

func (manager *Manager) scan() {
	now := time.Now()
	var overdueIDs []string
	var overdues []time.Duration

	{
		manager.mu.Lock()
		defer manager.mu.Unlock()

		for id, target := range manager.targets {
			// Check overdue
			if dt := time.Since(target.lastHeartbeat); dt > target.period {
				// Check firePeriod
				if fd := now.Sub(target.lastFire); fd > manager.firePeriod {
					overdueIDs = append(overdueIDs, id)
					overdues = append(overdues, dt)
				}
			}
		}
	}

	// Fire callbacks
	for i, id := range overdueIDs {
		od := overdues[i]
		manager.overdueCallback(id, od)

		t := manager.targets[id]
		t.lastFire = now
		manager.targets[id] = t
	}

	// Self heartbeat
	if time.Since(manager.selfLastHeartbeat) > manager.selfHeartbeatPeriod {
		manager.selfLastHeartbeat = now
		logger.Log.Info().Str("component", "WATCHDOG").Msgf("SELF HEARTBEAT (is alive)")
	}
}
