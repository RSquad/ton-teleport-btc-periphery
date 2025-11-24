package alerts

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/watchdog"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/config"
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
	config *config.ServicesConfig,
) (*AlertManager, error) {
	alertFactory := NewAlertFactory(contractAddrs, config)

	// Setup watchdog
	watchdog.Global().Watch("AlertManager", time.Duration(300)*time.Second)

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
		manager.mu.RLock()
		if enforcedState, ok := manager.alertStatesEnforced[alertName]; ok {
			state = enforcedState
		}
		manager.mu.RUnlock()

		if state == nil { // State is not enforced
			severity, description, values, err := alert.Check(manager.dataSource)
			if err != nil {
				description = Description(err.Error())
			}

			state = NewAlertState(
				alertName,
				severity,
				description,
				err,
				false,
				values,
			)
		}

		manager.UpdateState(state)

		// Send to the alert dispatcher
		var dispatcherError error = nil
		if state.Severity >= SEVERITY_OK {
			dispatcherError = manager.alertDispatcher.OnAlert(state)
		}

		manager.LogAlert(state, dispatcherError)
	}

	watchdog.Global().Heartbeat("AlertManager")
}

func (manager *AlertManager) LogAlert(state *AlertState, extErr error) {
	if state.LastErr != nil {
		logger.Log.Error().
			Str("Alert", state.Name).
			Err(state.LastErr).
			Msg("Alert finished work with error")
	} else {
		logger.Log.Debug().
			Str("Alert", state.Name).
			Msg(string(state.Description))
	}

	if extErr != nil {
		logger.Log.Error().
			Str("Alert", state.Name).
			Err(extErr).
			Msg("Alert finished work with error")
	}
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
	manager.alertStates[state.Name] = state
	manager.mu.Unlock()
}

func (manager *AlertManager) EnforceState(state *AlertState) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	// Verify name
	_, ok := manager.alerts[state.Name]
	if !ok {
		return fmt.Errorf("alert not found: '%s'", state.Name)
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
