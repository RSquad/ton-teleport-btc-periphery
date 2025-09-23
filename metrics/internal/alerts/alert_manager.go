package alerts

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
	"github.com/xssnick/tonutils-go/address"
)

type AlertManager struct {
	alertsFactory   *AlertFactory
	alerts          map[string]Alert
	dataSource      AlertDataSource
	alertDispatcher AlertDispatcher
	contractAddrs   map[string]*address.Address

	mu                  sync.RWMutex
	alertStates         map[string]*AlertState
	alertStatesEnforced map[string]*AlertState // for test purpose only
}

func NewAlertManager(
	dataSource AlertDataSource,
	alertDispatcher AlertDispatcher,
	contractAddrs map[string]*address.Address,
) (*AlertManager, error) {
	alertFactory := NewAlertFactory(contractAddrs)

	alertManager := AlertManager{
		alertsFactory:       alertFactory,
		alerts:              make(map[string]Alert),
		dataSource:          dataSource,
		alertDispatcher:     alertDispatcher,
		contractAddrs:       contractAddrs,
		alertStates:         make(map[string]*AlertState),
		alertStatesEnforced: make(map[string]*AlertState),
	}

	for name, factory := range alertFactory.Factories {
		alertManager.RegisterAlert(name, factory())
	}

	return &alertManager, nil
}

func (manager *AlertManager) RegisterAlert(name string, alert Alert) error {
	if _, exists := manager.alerts[name]; exists {
		return fmt.Errorf("Alert with name `%s` already exists", name)
	}

	manager.alerts[name] = alert

	return nil
}

func (manager *AlertManager) CheckAll() {
	for alertName, alert := range manager.alerts {
		var state *AlertState = nil
		// Check enforced state
		{
			manager.mu.RLock()
			if enforcedState, ok := manager.alertStatesEnforced[alertName]; ok {
				state = enforcedState
			}
			manager.mu.RUnlock()
		}

		if state == nil { // State in not enforced
			severity, labels, values, err := alert.Check(manager.dataSource)

			state = NewAlertState(
				alertName,
				severity,
				labels,
				err,
				false,
				values,
			)

			if err != nil {
				manager.UpdateState(state)
				manager.LogAlertError(alertName, err)
				continue
			}
		}

		manager.UpdateState(state)

		if state.Severity >= SEVERITY_OK {
			err := manager.alertDispatcher.OnAlert(state)
			if err != nil {
				manager.LogAlertError(alertName, err)
			}
		}
	}
}

func (manager *AlertManager) LogAlertError(alertName string, err error) {
	logger.Log.Error().
		Str("Alert", alertName).
		Err(err).
		Msg("Alert finished work with error")
}

func (manager *AlertManager) GetAlertState(name string) (AlertState, error) {
	manager.mu.RLock()
	defer manager.mu.RUnlock()

	alertState, ok := manager.alertStates[name]
	if !ok {
		return AlertState{}, fmt.Errorf("Alert not found for name '%s'", name)
	}

	return alertState.DeepCopy(), nil
}

func (manager *AlertManager) GetInfoJson() (string, error) {
	manager.mu.RLock()
	defer manager.mu.RUnlock()

	jsonData, err := json.Marshal(manager.alertStates)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

func (manager *AlertManager) GetEnforceInfoJsonStr() (string, error) {
	manager.mu.RLock()
	defer manager.mu.RUnlock()

	jsonData, err := json.Marshal(manager.alertStatesEnforced)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

func (manager *AlertManager) UpdateState(state *AlertState) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	manager.alertStates[state.Name] = state
}

func (manager *AlertManager) EnforceState(state *AlertState) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	// Verify name
	_, ok := manager.alerts[state.Name]
	if !ok {
		return fmt.Errorf("alert not found: '%s'", state.Name)
	}

	// Verify labels
	alert, err := manager.alertsFactory.NewAlertInstance(state.Name)
	if !ok {
		return err
	}

	expectedLabels := alert.NewLabels()
	expectedLabelsKeys := mutils.ExtractMapKeys(expectedLabels)
	expectedLabelsStr := mutils.JoinToStr(expectedLabelsKeys)
	stateLabelsKeys := mutils.ExtractMapKeys(state.Labels)
	stateLabelsStr := mutils.JoinToStr(stateLabelsKeys)

	if len(expectedLabelsKeys) != len(state.Labels) {
		return fmt.Errorf("expected labels '%s', but got '%s'", expectedLabelsStr, stateLabelsStr)
	}

	for name := range expectedLabels {
		_, ok := state.Labels[name]
		if !ok {
			return fmt.Errorf("expected labels '%s', but got '%s'", expectedLabelsStr, stateLabelsStr)
		}
	}

	// Update enforced state
	manager.alertStatesEnforced[state.Name] = state

	return nil
}

func (manager *AlertManager) ResetEnforceState(name string) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	_, ok := manager.alertStatesEnforced[name]
	if !ok {
		return fmt.Errorf("enforced alert not found: '%s'", name)
	}

	delete(manager.alertStatesEnforced, name)

	return nil
}
