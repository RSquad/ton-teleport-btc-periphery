package alerts

import (
	"maps"
	"time"
)

type AlertState struct {
	Name         string
	Severity     Severity
	Description  Description
	LastErr      error
	Enforced     bool
	LastUpdateTs time.Time
	Values       Values
}

func NewAlertState(
	name string,
	severity Severity,
	description Description,
	lastErr error,
	enforced bool,
	values Values,
) *AlertState {
	return &AlertState{
		Name:         name,
		Severity:     severity,
		Description:  description,
		LastErr:      lastErr,
		Enforced:     enforced,
		LastUpdateTs: time.Now(),
		Values:       values,
	}
}

func (state AlertState) DeepCopy() AlertState {
	var valuesCopy Values
	if state.Values != nil {
		valuesCopy = make(Values, len(state.Values))
		maps.Copy(valuesCopy, state.Values)
	}

	return AlertState{
		Name:         state.Name,
		Severity:     state.Severity,
		Description:  state.Description,
		LastErr:      state.LastErr,
		Enforced:     state.Enforced,
		LastUpdateTs: state.LastUpdateTs,
		Values:       valuesCopy,
	}
}
