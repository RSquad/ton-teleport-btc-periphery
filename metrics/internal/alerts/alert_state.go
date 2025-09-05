package alerts

import "time"

type AlertState struct {
	Name         string
	Severity     Severity
	Labels       Labels
	LastErr      error
	Enforced     bool
	LastUpdateTs time.Time
}

func NewAlertState(
	name string,
	severity Severity,
	labels Labels,
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

func (state AlertState) DeepCopy() AlertState {
	var labelsCopy Labels
	if state.Labels != nil {
		labelsCopy = make(Labels, len(state.Labels))
		for k, v := range state.Labels {
			labelsCopy[k] = v
		}
	}

	return AlertState{
		Name:         state.Name,
		Severity:     state.Severity,
		Labels:       labelsCopy,
		LastErr:      state.LastErr,
		Enforced:     state.Enforced,
		LastUpdateTs: state.LastUpdateTs,
	}
}
