package alerts

import (
	"database/sql"
	"errors"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/metrics"
)

type AlertDataSourceLive struct {
	metricsManager *metrics.MetricsManager
}

func NewAlertDataSourceLive(db *sql.DB, tonClient *tonclient.TonClient) *AlertDataSourceLive {
	dataSource := AlertDataSourceLive{
		metricsManager: metrics.NewMetricsManager(db, tonClient),
	}

	return &dataSource
}

func (dataSource *AlertDataSourceLive) FirstUnsignedPegout() (*coordinator.PegoutRecord, error) {
	coordinatorContractStateJson, err := dataSource.metricsManager.CoordinatorContractState()
	if err != nil {
		return nil, err
	}

	unsignedPegoutsJson, ok := coordinatorContractStateJson["UnsignedPegouts"].([]interface{})
	if !ok {
		return nil, errors.New("UnsignedPegouts has wrong type")
	}

	if len(unsignedPegoutsJson) == 0 {
		return nil, nil
	}

	unsignedPegouts, err := data_models.DeserializePegouts(unsignedPegoutsJson, 1)
	if err != nil {
		return nil, err
	}

	return &unsignedPegouts[0], nil
}

func (dataSource *AlertDataSourceLive) ConfiguratorContractData() (*coordinator.Storage, error) {
	coordinatorContractStateJson, err := dataSource.metricsManager.CoordinatorContractState()
	if err != nil {
		return nil, err
	}

	coordinatorContractState, err := data_models.DeserializeCoordinatorContractState(coordinatorContractStateJson)
	if err != nil {
		return nil, err
	}

	return coordinatorContractState, nil
}

func (dataSource *AlertDataSourceLive) Dkg() (*coordinator.DKG, error) {
	prevDkgJson, err := dataSource.metricsManager.Dkg()
	if err != nil {
		return nil, err
	}

	prevDkg, err := data_models.DeserializeDkg(prevDkgJson)
	if err != nil {
		return nil, err
	}

	return prevDkg, nil
}

func (dataSource *AlertDataSourceLive) PrevDkg() (*coordinator.DKG, error) {
	prevDkgJson, err := dataSource.metricsManager.PrevDkg()
	if err != nil {
		return nil, err
	}

	prevDkg, err := data_models.DeserializeDkg(prevDkgJson)
	if err != nil {
		return nil, err
	}

	return prevDkg, nil
}
