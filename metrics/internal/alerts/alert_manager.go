package alerts

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
)

type AlertManager struct {
	alerts          map[string]Alert
	dataSource      AlertDataSource
	alertDispatcher AlertDispatcher

	mu                  sync.RWMutex
	alertStates         map[string]*AlertState
	alertStatesEnforced map[string]*AlertState // for test purpose only
}

func NewAlertManager(
	dataSource AlertDataSource,
	alertDispatcher AlertDispatcher,
) (*AlertManager, error) {
	alertManager := AlertManager{
		alerts:              make(map[string]Alert),
		dataSource:          dataSource,
		alertDispatcher:     alertDispatcher,
		alertStates:         make(map[string]*AlertState),
		alertStatesEnforced: make(map[string]*AlertState),
	}

	var err error = nil

	// alert_pegout_fee_not_reset (pegout.fee.not.reset)
	err = alertManager.RegisterAlert(
		"alert_pegout_fee_not_reset",
		NewAlertPegoutFeeNotReset(),
	)
	if err != nil {
		return nil, err
	}

	// alert_pegout_signers (pegout.signers)
	err = alertManager.RegisterAlert(
		"alert_pegout_signers",
		NewAlertPegoutSigners(),
	)
	if err != nil {
		return nil, err
	}

	// alert_pegout_restarts (pegout.restarts)
	err = alertManager.RegisterAlert(
		"alert_pegout_restarts",
		NewAlertPegoutRestarts(),
	)
	if err != nil {
		return nil, err
	}

	// alert_pegout_commintments (pegout.commitments)
	err = alertManager.RegisterAlert(
		"alert_pegout_commintments",
		NewAlertPegoutCommintments(),
	)
	if err != nil {
		return nil, err
	}

	// alert_pegout_in_mempool (pegout.in.mempool)
	err = alertManager.RegisterAlert(
		"alert_pegout_in_mempool",
		NewAlertPegoutInMempool(),
	)
	if err != nil {
		return nil, err
	}

	// alert_pegout_signing_duration (pegout.signing.duration)
	err = alertManager.RegisterAlert(
		"alert_pegout_signing_duration",
		NewAlertPegoutSigningDuration(),
	)
	if err != nil {
		return nil, err
	}

	// dkg_status (dkg.status)
	err = alertManager.RegisterAlert(
		"dkg_status",
		NewAlertDkgStatus(),
	)
	if err != nil {
		return nil, err
	}

	// fees_health (fees.health)
	// TODO: add

	// dkg_restarts (dkg.restarts)
	err = alertManager.RegisterAlert(
		"dkg_restarts",
		NewAlertDkgRestarts(),
	)
	if err != nil {
		return nil, err
	}

	// dkg_participants (dkg.participants)
	err = alertManager.RegisterAlert(
		"dkg_participants",
		NewAlertDkgParticipants(),
	)
	if err != nil {
		return nil, err
	}

	// dkg_culprit_found (dkg.culprit.found)
	err = alertManager.RegisterAlert(
		"dkg_culprit_found",
		NewAlertDkgCulpritFound(),
	)
	if err != nil {
		return nil, err
	}

	// contract_balance_coordinator (contract.balance.coordinator)
	err = alertManager.RegisterAlert(
		"contract_balance_coordinator",
		NewAlertContractBalance("coordinator"),
	)
	if err != nil {
		return nil, err
	}

	// contract_balance_teleport (contract.balance.teleport)
	err = alertManager.RegisterAlert(
		"contract_balance_teleport",
		NewAlertContractBalance("teleport"),
	)
	if err != nil {
		return nil, err
	}

	// contract_balance_bitclient (contract.balance.bitclient)
	err = alertManager.RegisterAlert(
		"contract_balance_bitclient",
		NewAlertContractBalance("bitclient"),
	)
	if err != nil {
		return nil, err
	}

	// contract_balance_minter (contract.balance.minter)
	err = alertManager.RegisterAlert(
		"contract_balance_minter",
		NewAlertContractBalance("minter"),
	)
	if err != nil {
		return nil, err
	}

	// contract_balance_relayer (contract.balance.relayer)
	err = alertManager.RegisterAlert(
		"contract_balance_relayer",
		NewAlertContractBalance("relayer"),
	)
	if err != nil {
		return nil, err
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
			severity, labels, err := alert.Check(manager.dataSource)

			state = NewAlertState(
				alertName,
				severity,
				labels,
				err,
				false,
			)

			if err != nil {
				manager.UpdateState(state)
				manager.LogAlertError(alertName, err)
				continue
			}
		}

		manager.UpdateState(state)

		if state.Severity >= SEVERITY_OK {
			manager.alertDispatcher.OnAlert(state)
		}
	}
}

func (manager *AlertManager) LogAlertError(alertName string, err error) {
	logger.Log.Error().
		Str("Alert", alertName).
		Err(err).
		Msg("Alert finished work with error")
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

func (manager *AlertManager) EnforceState(state *AlertState) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	manager.alertStatesEnforced[state.Name] = state
}

func (manager *AlertManager) ResetEnforceState(name string) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	delete(manager.alertStatesEnforced, name)
}
