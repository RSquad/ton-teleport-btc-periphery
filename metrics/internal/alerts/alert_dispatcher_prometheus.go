package alerts

import (
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
	gv, ok := d.gauges[state.Name]
	if ok {
		return gv
	}

	if gv, ok = d.gauges[state.Name]; ok {
		return gv
	}

	labelNames := []string{"description"}

	gv = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: state.Name,
		},
		labelNames,
	)
	d.gauges[state.Name] = gv
	return gv
}

func (d *AlertDispatcherPrometheus) OnAlert(state *AlertState) error {
	if state == nil {
		return nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	gv := d.getOrCreateGaugeVec(state)
	gv.Reset()

	g, err := gv.GetMetricWith(prometheus.Labels{"description": ""})
	if err != nil {
		return err
	}
	g.Set(float64(state.Severity))

	return nil
}
