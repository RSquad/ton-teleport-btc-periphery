package alerts

import (
	"context"
	"errors"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
	"github.com/xssnick/tonutils-go/address"
)

type (
	CoordinatorContractStorageDbFn                func() (*coordinator.Storage, error)
	TeleportContractStorageDbFn                   func() (*teleportcontract.Storage, error)
	BitcoinClientContractStorageDbFn              func() (*data_models.BitcoinClientContractStorage, error)
	FirstUnsignedPegoutDbFn                       func() (*coordinator.PegoutRecord, error)
	LastConfirmedPegout                           func() (*data_models.Pegout, error)
	LastSignedPegoutDbFn                          func() (*data_models.Pegout, error)
	LastSignedPegoutsDbFn                         func(limit uint) ([]*data_models.Pegout, error)
	PegoutDbFn                                    func(address *address.Address) (*data_models.Pegout, error)
	DkgDbFn                                       func() (*coordinator.DKG, error)
	PrevDkgDbFn                                   func() (*coordinator.DKG, error)
	DkgBeforeRestartDbFn                          func(t time.Time) (*coordinator.DKG, error)
	BtcGetBlockHashByTxIdFn                       func(txID *chainhash.Hash) (*chainhash.Hash, error)
	BtcGetBlockHeightByHashFn                     func(hash *chainhash.Hash) (int64, error)
	BtcGetCpfpLengthFn                            func(hash *chainhash.Hash) (int, error)
	BtcGetMempoolEntryFn                          func(txHash string) (*btcjson.GetMempoolEntryResult, error)
	BtcGetBestBlockHeightFn                       func() (int, error)
	BitcoinClientContractLastConfirmedBlockHashFn func() (*chainhash.Hash, error)
	TonMaxMainValidatorsFn                        func(ctx context.Context) (int, error)
	ActualContractBalanceFn                       func(name string) (int64, error)
	NowUnixTsFn                                   func() int64
)

// Config holds optional function callbacks
type AlertDataSourceTestingConfig struct {
	CoordinatorContractStorageDbFn                CoordinatorContractStorageDbFn
	TeleportContractStorageDbFn                   TeleportContractStorageDbFn
	BitcoinClientContractStorageDbFn              BitcoinClientContractStorageDbFn
	FirstUnsignedPegoutDbFn                       FirstUnsignedPegoutDbFn
	LastConfirmedPegoutFn                         LastConfirmedPegout
	LastSignedPegoutDbFn                          LastSignedPegoutDbFn
	LastSignedPegoutsDbFn                         LastSignedPegoutsDbFn
	PegoutDbFn                                    PegoutDbFn
	DkgDbFn                                       DkgDbFn
	PrevDkgDbFn                                   PrevDkgDbFn
	DkgBeforeRestartDbFn                          DkgBeforeRestartDbFn
	BtcGetBlockHashByTxIdFn                       BtcGetBlockHashByTxIdFn
	BtcGetBlockHeightByHashFn                     BtcGetBlockHeightByHashFn
	BtcGetCpfpLengthFn                            BtcGetCpfpLengthFn
	BtcGetMempoolEntryFn                          BtcGetMempoolEntryFn
	BtcGetBestBlockHeightFn                       BtcGetBestBlockHeightFn
	BitcoinClientContractLastConfirmedBlockHashFn BitcoinClientContractLastConfirmedBlockHashFn
	TonMaxMainValidatorsFn                        TonMaxMainValidatorsFn
	ActualContractBalanceFn                       ActualContractBalanceFn
	NowUnixTsFn                                   NowUnixTsFn
}

type AlertDataSourceTesting struct {
	cfg AlertDataSourceTestingConfig
}

func NewAlertDataSourceTesting(cfg AlertDataSourceTestingConfig) AlertDataSource {
	return &AlertDataSourceTesting{cfg: cfg}
}

func (dataSource *AlertDataSourceTesting) CoordinatorContractStorageDB() (*coordinator.Storage, error) {
	if dataSource.cfg.CoordinatorContractStorageDbFn == nil {
		return nil, errors.New("CoordinatorContractStorageDbFn callback not set")
	}
	return dataSource.cfg.CoordinatorContractStorageDbFn()
}

func (dataSource *AlertDataSourceTesting) TeleportContractStorageDB() (*teleportcontract.Storage, error) {
	if dataSource.cfg.TeleportContractStorageDbFn == nil {
		return nil, errors.New("TeleportContractStorageDbFn callback not set")
	}
	return dataSource.cfg.TeleportContractStorageDbFn()
}

func (dataSource *AlertDataSourceTesting) BitcoinClientContractStorageDB() (*data_models.BitcoinClientContractStorage, error) {
	if dataSource.cfg.BitcoinClientContractStorageDbFn == nil {
		return nil, errors.New("BitcoinClientContractStorageDbFn callback not set")
	}
	return dataSource.cfg.BitcoinClientContractStorageDbFn()
}

func (dataSource *AlertDataSourceTesting) FirstUnsignedPegoutDB() (*coordinator.PegoutRecord, error) {
	if dataSource.cfg.FirstUnsignedPegoutDbFn == nil {
		return nil, errors.New("FirstUnsignedPegoutDbFn callback not set")
	}
	return dataSource.cfg.FirstUnsignedPegoutDbFn()
}

func (dataSource *AlertDataSourceTesting) LastConfirmedPegout() (*data_models.Pegout, error) {
	if dataSource.cfg.LastConfirmedPegoutFn == nil {
		return nil, errors.New("LastConfirmedPegout callback not set")
	}
	return dataSource.cfg.LastConfirmedPegoutFn()
}

func (dataSource *AlertDataSourceTesting) LastSignedPegoutDB() (*data_models.Pegout, error) {
	if dataSource.cfg.LastSignedPegoutDbFn == nil {
		return nil, errors.New("LastSignedPegoutDbFn callback not set")
	}
	return dataSource.cfg.LastSignedPegoutDbFn()
}

func (dataSource *AlertDataSourceTesting) LastSignedPegoutsDB(limit uint) ([]*data_models.Pegout, error) {
	if dataSource.cfg.LastSignedPegoutsDbFn == nil {
		return nil, errors.New("LastSignedPegoutsDbFn callback not set")
	}
	return dataSource.cfg.LastSignedPegoutsDbFn(limit)
}

func (dataSource *AlertDataSourceTesting) DkgDB() (*coordinator.DKG, error) {
	if dataSource.cfg.DkgDbFn == nil {
		return nil, errors.New("DkgDbFn callback not set")
	}
	return dataSource.cfg.DkgDbFn()
}

func (dataSource *AlertDataSourceTesting) PrevDkgDB() (*coordinator.DKG, error) {
	if dataSource.cfg.PrevDkgDbFn == nil {
		return nil, errors.New("PrevDkgDbFn callback not set")
	}
	return dataSource.cfg.PrevDkgDbFn()
}

func (dataSource *AlertDataSourceTesting) DkgBeforeRestartDB(t time.Time) (*coordinator.DKG, error) {
	if dataSource.cfg.DkgBeforeRestartDbFn == nil {
		return nil, errors.New("DkgBeforeRestartDbFn callback not set")
	}
	return dataSource.cfg.DkgBeforeRestartDbFn(t)
}

func (dataSource *AlertDataSourceTesting) PegoutDB(address *address.Address) (*data_models.Pegout, error) {
	if dataSource.cfg.PegoutDbFn == nil {
		return nil, errors.New("PegoutDbFn callback not set")
	}
	return dataSource.cfg.PegoutDbFn(address)
}

func (dataSource *AlertDataSourceTesting) BtcGetBlockHashByTxID(txID *chainhash.Hash) (*chainhash.Hash, error) {
	if dataSource.cfg.BtcGetBlockHashByTxIdFn == nil {
		return nil, errors.New("BtcGetBlockHashByTxIdFn callback not set")
	}
	return dataSource.cfg.BtcGetBlockHashByTxIdFn(txID)
}

func (dataSource *AlertDataSourceTesting) BtcGetBlockHeightByHash(hash *chainhash.Hash) (int64, error) {
	if dataSource.cfg.BtcGetBlockHeightByHashFn == nil {
		return 0, errors.New("BtcGetBlockHeightByHashFn callback not set")
	}
	return dataSource.cfg.BtcGetBlockHeightByHashFn(hash)
}

func (dataSource *AlertDataSourceTesting) BtcGetBestBlockHeight() (int, error) {
	if dataSource.cfg.BtcGetBestBlockHeightFn == nil {
		return 0, errors.New("BtcGetBestBlockHeightFn callback not set")
	}
	return dataSource.cfg.BtcGetBestBlockHeightFn()
}

func (dataSource *AlertDataSourceTesting) BtcGetCpfpLength(hash *chainhash.Hash) (int, error) {
	if dataSource.cfg.BtcGetCpfpLengthFn == nil {
		return 0, errors.New("BtcGetCpfpLengthFn callback not set")
	}
	return dataSource.cfg.BtcGetCpfpLengthFn(hash)
}

func (dataSource *AlertDataSourceTesting) BtcGetMempoolEntry(txHash string) (*btcjson.GetMempoolEntryResult, error) {
	if dataSource.cfg.BtcGetMempoolEntryFn == nil {
		return nil, errors.New("BtcGetMempoolEntryFn callback not set")
	}
	return dataSource.cfg.BtcGetMempoolEntryFn(txHash)
}

func (dataSource *AlertDataSourceTesting) BitcoinClientContractLastConfirmedBlockHash() (*chainhash.Hash, error) {
	if dataSource.cfg.BitcoinClientContractLastConfirmedBlockHashFn == nil {
		return nil, errors.New("BitcoinClientContractLastConfirmedBlockHashFn callback not set")
	}
	return dataSource.cfg.BitcoinClientContractLastConfirmedBlockHashFn()
}

func (dataSource *AlertDataSourceTesting) TonMaxMainValidators(ctx context.Context) (int, error) {
	if dataSource.cfg.TonMaxMainValidatorsFn == nil {
		return 0, errors.New("TonMaxMainValidatorsFn callback not set")
	}
	return dataSource.cfg.TonMaxMainValidatorsFn(ctx)
}

func (dataSource *AlertDataSourceTesting) ActualContractBalance(name string) (int64, error) {
	if dataSource.cfg.ActualContractBalanceFn == nil {
		return 0, errors.New("ActualContractBalanceFn callback not set")
	}
	return dataSource.cfg.ActualContractBalanceFn(name)
}

func (dataSource *AlertDataSourceTesting) NowUnixTs() int64 {
	return dataSource.cfg.NowUnixTsFn()
}
