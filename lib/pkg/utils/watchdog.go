package utils

import (
	"context"
	"errors"
	"sync"
	"time"
)

type WatchdogOverdueCallback func(id string, overdue time.Duration)

type Watchdog struct {
	period          time.Duration
	overdueCallback WatchdogOverdueCallback

	mu            sync.Mutex
	lastHeartbeat map[string]time.Time
	closed        chan struct{}
	wg            sync.WaitGroup
}

func NewWatchdog(period time.Duration, overdueCallback WatchdogOverdueCallback) (*Watchdog, error) {
	if period <= 0 {
		return nil, errors.New("watchdog period must be > 0")
	}

	if overdueCallback == nil {
		return nil, errors.New("watchdog overdue callback is null")
	}

	return &Watchdog{
		period:          period,
		overdueCallback: overdueCallback,
		lastHeartbeat:   make(map[string]time.Time),
		closed:          make(chan struct{}),
	}, nil
}

func (w *Watchdog) Start(ctx context.Context) {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		t := time.NewTicker(w.period / 2)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-w.closed:
				return
			case <-t.C:
				w.scan()
			}
		}
	}()
}

func (w *Watchdog) Stop() {
	close(w.closed)
	w.wg.Wait()
}

func (w *Watchdog) Watch(id string) {
	now := time.Now()
	w.mu.Lock()
	w.lastHeartbeat[id] = now
	w.mu.Unlock()
}

func (w *Watchdog) Unwatch(id string) {
	w.mu.Lock()
	delete(w.lastHeartbeat, id)
	w.mu.Unlock()
}

func (w *Watchdog) Heartbeat(id string) {
	now := time.Now()
	w.mu.Lock()
	if _, ok := w.lastHeartbeat[id]; ok {
		w.lastHeartbeat[id] = now
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
			if d := now.Sub(last); d > w.period {
				overdueIDs = append(overdueIDs, id)
				overdues = append(overdues, d)
			}
		}
	}

	// Fire callbacks
	for i, id := range overdueIDs {
		od := overdues[i]
		go w.overdueCallback(id, od)
	}
}
