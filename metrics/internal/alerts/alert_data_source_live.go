package alerts

import (
	"database/sql"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_sources"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/metrics"
	"github.com/xssnick/tonutils-go/address"
)

type AlertDataSourceLive struct {
	metricsManager *metrics.MetricsManager
	dataSourceDB   *data_sources.DataSourceDB
}

func NewAlertDataSourceLive(db *sql.DB, tonClient *tonclient.TonClient) AlertDataSource {
	dataSource := AlertDataSourceLive{
		metricsManager: metrics.NewMetricsManager(db, tonClient),
		dataSourceDB:   data_sources.NewDataSourceDB(db),
	}

	return &dataSource
}

func (dataSource *AlertDataSourceLive) CoordinatorContractDataDB() (*coordinator.Storage, error) {
	return dataSource.dataSourceDB.CoordinatorContractStorage()
}

func (dataSource *AlertDataSourceLive) FirstUnsignedPegoutDB() (*coordinator.PegoutRecord, error) {
	coordinatorContractStorage, err := dataSource.dataSourceDB.CoordinatorContractStorage()
	if err != nil {
		return nil, err
	}

	if len(coordinatorContractStorage.UnsignedPegouts) == 0 {
		return nil, nil
	}

	return &coordinatorContractStorage.UnsignedPegouts[0], nil
}

func (dataSource *AlertDataSourceLive) LastSignedPegoutDB() (*data_models.PegoutDbRow, error) {
	return dataSource.dataSourceDB.LastSignedPegoutDbRow()
}

func (dataSource *AlertDataSourceLive) DkgDB() (*coordinator.DKG, error) {
	return dataSource.dataSourceDB.Dkg()
}

func (dataSource *AlertDataSourceLive) PrevDkgDB() (*coordinator.DKG, error) {
	return dataSource.dataSourceDB.PrevDkg()
}

func (dataSource *AlertDataSourceLive) PegoutDB(address *address.Address) (*data_models.PegoutDbRow, error) {
	return dataSource.dataSourceDB.PegoutDbRow(address)
}

func (dataSource *AlertDataSourceLive) NowUnixTs() int64 {
	return time.Now().Unix()
}
