package alerts

import (
	"maps"
	"time"
)

type AlertState struct {
	Name         string // alert ID
	Severity     Severity
	Description  Description
	LastErr      error
	Enforced     bool
	FirstSeen    time.Time
	LastUpdateTs time.Time
	Values       Values
	IsActive     bool
	RepeatCount  int
	Hash         string
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
		FirstSeen:    time.Now(),
		LastUpdateTs: time.Now(),
		Values:       values,
		RepeatCount:  0,
		IsActive:     true,
		Hash:         "",
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
		FirstSeen:    state.FirstSeen,
		IsActive:     state.IsActive,
		RepeatCount:  state.RepeatCount,
		Hash:         state.Hash,
	}
}
