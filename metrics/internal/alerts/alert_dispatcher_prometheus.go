package alerts

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type AlertDispatcherPrometheus struct {
	mu         sync.RWMutex
	gauges     map[string]*prometheus.GaugeVec
	lastValues map[string]*AlertState
}

func NewAlertDispatcherPrometheus() *AlertDispatcherPrometheus {
	return &AlertDispatcherPrometheus{
		gauges:     make(map[string]*prometheus.GaugeVec),
		lastValues: make(map[string]*AlertState),
	}
}

func (d *AlertDispatcherPrometheus) getOrCreateGaugeVec(state *AlertState) *prometheus.GaugeVec {
	d.mu.RLock()
	gv, ok := d.gauges[state.Name]
	d.mu.RUnlock()
	if ok {
		return gv
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if gv, ok = d.gauges[state.Name]; ok {
		return gv
	}

	labelNames := make([]string, 0)
	for name := range state.Labels {
		labelNames = append(labelNames, name)
	}

	gv = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: state.Name,
			Help: fmt.Sprintf("Severity gauge for `%s`", state.Name),
		},
		labelNames,
	)
	d.gauges[state.Name] = gv
	return gv
}

func (d *AlertDispatcherPrometheus) OnAlert(state *AlertState) {
	gv := d.getOrCreateGaugeVec(state)

	d.mu.Lock()
	defer d.mu.Unlock()

	// Update vector value
	gv.With(prometheus.Labels(state.Labels)).Set(float64(state.Severity))

	// Remove the old value if the labels do not match
	lastValue, exists := d.lastValues[state.Name]
	if exists {
		if !IsEqual(lastValue.Labels, state.Labels) {
			gv.Delete(prometheus.Labels(lastValue.Labels))
		}
	}

	// Save last value
	d.lastValues[state.Name] = state
}
