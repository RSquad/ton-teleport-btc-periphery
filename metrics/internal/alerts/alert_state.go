package alerts

import "time"

type AlertState struct {
	Name         string
	Severity     Severity
	Labels       []string
	LastErr      error
	Enforced     bool
	LastUpdateTs time.Time
}

func NewAlertState(
	name string,
	severity Severity,
	labels []string,
	lastErr error,
	enforced bool,
) *AlertState {
	return &AlertState{
		Name:         name,
		Severity:     severity,
		Labels:       labels,
		LastErr:      lastErr,
		Enforced:     enforced,
		LastUpdateTs: time.Now(),
	}
}
