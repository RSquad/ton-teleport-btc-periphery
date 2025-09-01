package alerts

import (
	"database/sql"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_sources"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/metrics"
	"github.com/xssnick/tonutils-go/address"
)

type AlertDataSourceLive struct {
	metricsManager *metrics.MetricsManager
	dataSourceDB   *data_sources.DataSourceDB
	bitcoinClient  *bitcoin.Client
}

func NewAlertDataSourceLive(
	db *sql.DB,
	tonClient *tonclient.TonClient,
	bitcoinClient *bitcoin.Client,
) AlertDataSource {
	dataSource := AlertDataSourceLive{
		metricsManager: metrics.NewMetricsManager(db, tonClient),
		dataSourceDB:   data_sources.NewDataSourceDB(db),
		bitcoinClient:  bitcoinClient,
	}

	return &dataSource
}

func (dataSource *AlertDataSourceLive) CoordinatorContractStorageDB() (*coordinator.Storage, error) {
	return dataSource.dataSourceDB.CoordinatorContractStorage()
}

func (dataSource *AlertDataSourceLive) TeleportContractStorageDB() (*teleportcontract.Storage, error) {
	return dataSource.dataSourceDB.TeleportContractStorage()
}

func (dataSource *AlertDataSourceLive) BitcoinClientContractStorageDB() (*data_models.BitcoinClientContractStorage, error) {
	return dataSource.dataSourceDB.BitcoinClientContractStorage()
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

func (dataSource *AlertDataSourceLive) LastSignedPegoutDB() (*data_models.Pegout, error) {
	return dataSource.dataSourceDB.LastSignedPegout()
}

func (dataSource *AlertDataSourceLive) DkgDB() (*coordinator.DKG, error) {
	return dataSource.dataSourceDB.Dkg()
}

func (dataSource *AlertDataSourceLive) PrevDkgDB() (*coordinator.DKG, error) {
	return dataSource.dataSourceDB.PrevDkg()
}

func (dataSource *AlertDataSourceLive) PegoutDB(address *address.Address) (*data_models.Pegout, error) {
	return dataSource.dataSourceDB.Pegout(address)
}

func (dataSource *AlertDataSourceLive) NowUnixTs() int64 {
	return time.Now().Unix()
}

func (dataSource *AlertDataSourceLive) BtcGetBlockHashByTxID(txID *chainhash.Hash) (*chainhash.Hash, error) {
	return dataSource.bitcoinClient.GetBlockHashByTxID(txID)
}

func (dataSource *AlertDataSourceLive) BtcGetBlockHeightByHash(hash *chainhash.Hash) (int64, error) {
	return dataSource.bitcoinClient.GetBlockHeightByHash(hash)
}
