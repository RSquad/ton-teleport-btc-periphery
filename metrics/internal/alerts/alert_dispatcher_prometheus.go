package alerts

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type AlertDispatcherPrometheus struct {
	mu     sync.RWMutex
	gauges map[string]*prometheus.GaugeVec
}

func NewAlertDispatcherPrometheus() *AlertDispatcherPrometheus {
	return &AlertDispatcherPrometheus{
		gauges: make(map[string]*prometheus.GaugeVec),
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
	for name, _ := range state.Labels {
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
	gv.With(prometheus.Labels(state.Labels)).Set(float64(state.Severity))
}
