package alerts

import (
	"errors"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
	"github.com/xssnick/tonutils-go/address"
)

type (
	CoordinatorContractStorageDbFn              func() (*coordinator.Storage, error)
	TeleportContractStorageDbFn                 func() (*teleportcontract.Storage, error)
	BitcoinClientContractStorageDbFn            func() (*data_models.BitcoinClientContractStorage, error)
	FirstUnsignedPegoutDbFn                     func() (*coordinator.PegoutRecord, error)
	LastSignedPegoutDbFn                        func() (*data_models.Pegout, error)
	PegoutDbFn                                  func(address *address.Address) (*data_models.Pegout, error)
	DkgDbFn                                     func() (*coordinator.DKG, error)
	PrevDkgDbFn                                 func() (*coordinator.DKG, error)
	BtcGetBlockHashByTxID                       func(txID *chainhash.Hash) (*chainhash.Hash, error)
	BtcGetBlockHeightByHash                     func(hash *chainhash.Hash) (int64, error)
	BitcoinClientContractLastConfirmedBlockHash func() (*chainhash.Hash, error)
	NowUnixTsFn                                 func() int64
)

// Config holds optional function callbacks
type AlertDataSourceTestingConfig struct {
	CoordinatorContractStorageDbFn                CoordinatorContractStorageDbFn
	TeleportContractStorageDbFn                   TeleportContractStorageDbFn
	BitcoinClientContractStorageDbFn              BitcoinClientContractStorageDbFn
	FirstUnsignedPegoutDbFn                       FirstUnsignedPegoutDbFn
	LastSignedPegoutDbFn                          LastSignedPegoutDbFn
	PegoutDbFn                                    PegoutDbFn
	DkgDbFn                                       DkgDbFn
	PrevDkgDbFn                                   PrevDkgDbFn
	BtcGetBlockHashByTxIdFn                       BtcGetBlockHashByTxID
	BtcGetBlockHeightByHashFn                     BtcGetBlockHeightByHash
	BitcoinClientContractLastConfirmedBlockHashFn BitcoinClientContractLastConfirmedBlockHash
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

func (dataSource *AlertDataSourceTesting) LastSignedPegoutDB() (*data_models.Pegout, error) {
	if dataSource.cfg.LastSignedPegoutDbFn == nil {
		return nil, errors.New("LastSignedPegoutDbFn callback not set")
	}
	return dataSource.cfg.LastSignedPegoutDbFn()
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

func (dataSource *AlertDataSourceTesting) BitcoinClientContractLastConfirmedBlockHash() (*chainhash.Hash, error) {
	if dataSource.cfg.BitcoinClientContractLastConfirmedBlockHashFn == nil {
		return nil, errors.New("BitcoinClientContractLastConfirmedBlockHashFn callback not set")
	}
	return dataSource.cfg.BitcoinClientContractLastConfirmedBlockHashFn()
}

func (dataSource *AlertDataSourceTesting) NowUnixTs() int64 {
	return dataSource.cfg.NowUnixTsFn()
}
