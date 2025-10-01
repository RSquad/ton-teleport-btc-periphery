package alerts

import (
	"maps"
	"time"
)

type AlertState struct {
	Name         string
	Severity     Severity
	Labels       Labels
	LastErr      error
	Enforced     bool
	LastUpdateTs time.Time
	Values       Values
}

func NewAlertState(
	name string,
	severity Severity,
	labels Labels,
	lastErr error,
	enforced bool,
	values Values,
) *AlertState {
	return &AlertState{
		Name:         name,
		Severity:     severity,
		Labels:       labels,
		LastErr:      lastErr,
		Enforced:     enforced,
		LastUpdateTs: time.Now(),
		Values:       values,
	}
}

func (state AlertState) DeepCopy() AlertState {
	var labelsCopy Labels
	if state.Labels != nil {
		labelsCopy = make(Labels, len(state.Labels))
		maps.Copy(labelsCopy, state.Labels)
	}

	var valuesCopy Values
	if state.Values != nil {
		valuesCopy = make(Values, len(state.Values))
		maps.Copy(valuesCopy, state.Values)
	}

	return AlertState{
		Name:         state.Name,
		Severity:     state.Severity,
		Labels:       labelsCopy,
		LastErr:      state.LastErr,
		Enforced:     state.Enforced,
		LastUpdateTs: state.LastUpdateTs,
		Values:       valuesCopy,
	}
}
