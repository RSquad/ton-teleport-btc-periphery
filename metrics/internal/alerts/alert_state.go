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
	IntValues    IntValues
}

func NewAlertState(
	name string,
	severity Severity,
	labels Labels,
	lastErr error,
	enforced bool,
	intValues IntValues,
) *AlertState {
	return &AlertState{
		Name:         name,
		Severity:     severity,
		Labels:       labels,
		LastErr:      lastErr,
		Enforced:     enforced,
		LastUpdateTs: time.Now(),
		IntValues:    intValues,
	}
}

func (state AlertState) DeepCopy() AlertState {
	var labelsCopy Labels
	if state.Labels != nil {
		labelsCopy = make(Labels, len(state.Labels))
		maps.Copy(labelsCopy, state.Labels)
	}

	var intValuesCopy IntValues
	if state.IntValues != nil {
		intValuesCopy = make(IntValues, len(state.IntValues))
		maps.Copy(intValuesCopy, state.IntValues)
	}

	return AlertState{
		Name:         state.Name,
		Severity:     state.Severity,
		Labels:       labelsCopy,
		LastErr:      state.LastErr,
		Enforced:     state.Enforced,
		LastUpdateTs: state.LastUpdateTs,
		IntValues:    intValuesCopy,
	}
}
