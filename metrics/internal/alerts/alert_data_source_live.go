package alerts

import (
	"context"
	"database/sql"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/config"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_sources"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"

	"github.com/xssnick/tonutils-go/address"
)

type AlertDataSourceLive struct {
	dataSourceDB        *data_sources.DataSourceDB
	bitcoinClient       *bitcoin.Client
	globalRuntimeConfig *config.GlobalRuntimeConfig
}

func NewAlertDataSourceLive(
	db *sql.DB,
	bitcoinClient *bitcoin.Client,
	globalRuntimeConfig *config.GlobalRuntimeConfig,
	contractAddrs map[string]*address.Address,
) AlertDataSource {
	dataSource := AlertDataSourceLive{
		dataSourceDB:        data_sources.NewDataSourceDB(db),
		bitcoinClient:       bitcoinClient,
		globalRuntimeConfig: globalRuntimeConfig,
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

	if coordinatorContractStorage == nil {
		return nil, nil
	}

	if len(coordinatorContractStorage.UnsignedPegouts) == 0 {
		return nil, nil
	}

	return &coordinatorContractStorage.UnsignedPegouts[0], nil
}

func (dataSource *AlertDataSourceLive) LastSignedPegoutDB() (*data_models.Pegout, error) {
	return dataSource.dataSourceDB.LastSignedPegout()
}

func (dataSource *AlertDataSourceLive) LastSignedPegoutsDB(limit uint) ([]*data_models.Pegout, error) {
	return dataSource.dataSourceDB.LastSignedPegouts(limit)
}

func (dataSource *AlertDataSourceLive) DkgDB() (*coordinator.DKG, error) {
	return dataSource.dataSourceDB.Dkg()
}

func (dataSource *AlertDataSourceLive) PrevDkgDB() (*coordinator.DKG, error) {
	return dataSource.dataSourceDB.PrevDkg()
}

func (dataSource *AlertDataSourceLive) DkgBeforeRestartDB(t time.Time) (*coordinator.DKG, error) {
	return dataSource.dataSourceDB.DkgBeforeRestart(t)
}

func (dataSource *AlertDataSourceLive) PegoutDB(address *address.Address) (*data_models.Pegout, error) {
	return dataSource.dataSourceDB.Pegout(address)
}

func (dataSource *AlertDataSourceLive) NowUnixTs() int64 {
	return time.Now().Unix()
}

func (dataSource *AlertDataSourceLive) BtcGetBestBlockHeight() (int, error) {
	return mutils.GetBestBlockHeight(dataSource.bitcoinClient)
}

func (dataSource *AlertDataSourceLive) BtcGetBlockHashByTxID(txID *chainhash.Hash) (*chainhash.Hash, error) {
	return dataSource.bitcoinClient.GetBlockHashByTxID(txID)
}

func (dataSource *AlertDataSourceLive) BtcGetBlockHeightByHash(hash *chainhash.Hash) (int64, error) {
	return dataSource.bitcoinClient.GetBlockHeightByHash(hash)
}

func (dataSource *AlertDataSourceLive) BtcGetMempoolEntry(txHash string) (*btcjson.GetMempoolEntryResult, error) {
	return dataSource.bitcoinClient.RPCClient.GetMempoolEntry(txHash)
}

func (dataSource *AlertDataSourceLive) TonMaxMainValidators(ctx context.Context) (int, error) {
	return dataSource.globalRuntimeConfig.TonMaxMainValidators(ctx)
}

func (dataSource *AlertDataSourceLive) ActualContractBalance(name string) (int64, error) {
	return dataSource.dataSourceDB.ActualContractBalance(name)
}
