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

func (d *AlertDispatcherPrometheus) getOrCreateGaugeVec(name string, labels []string) *prometheus.GaugeVec {
	d.mu.RLock()
	gv, ok := d.gauges[name]
	d.mu.RUnlock()
	if ok {
		return gv
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if gv, ok = d.gauges[name]; ok {
		return gv
	}

	gv = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name,
			Help: fmt.Sprintf("Severity gauge for `%s`", name),
		},
		labels,
	)
	d.gauges[name] = gv
	return gv
}

func (d *AlertDispatcherPrometheus) OnAlert(name string, labels []string, severity Severity) {
	gv := d.getOrCreateGaugeVec(name, labels)
	gv.WithLabelValues().Set(float64(severity))
}
