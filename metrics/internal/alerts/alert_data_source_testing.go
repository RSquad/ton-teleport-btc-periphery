package alerts

import (
	"errors"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
	"github.com/xssnick/tonutils-go/address"
)

type (
	CoordinatorContractDataDbFn func() (*coordinator.Storage, error)
	FirstUnsignedPegoutDbFn     func() (*coordinator.PegoutRecord, error)
	LastSignedPegoutDbFn        func() (*data_models.PegoutDbRow, error)
	PegoutDbFn                  func(address *address.Address) (*data_models.PegoutDbRow, error)
	DkgDbFn                     func() (*coordinator.DKG, error)
	PrevDkgDbFn                 func() (*coordinator.DKG, error)
	NowUnixTsFn                 func() int64
)

// Config holds optional function callbacks
type AlertDataSourceTestingConfig struct {
	CoordinatorContractDataDbFn CoordinatorContractDataDbFn
	FirstUnsignedPegoutDbFn     FirstUnsignedPegoutDbFn
	LastSignedPegoutDbFn        LastSignedPegoutDbFn
	PegoutDbFn                  PegoutDbFn
	DkgDbFn                     DkgDbFn
	PrevDkgDbFn                 PrevDkgDbFn
	NowUnixTsFn                 NowUnixTsFn
}

type AlertDataSourceTesting struct {
	cfg AlertDataSourceTestingConfig
}

func NewAlertDataSourceTesting(cfg AlertDataSourceTestingConfig) AlertDataSource {
	return &AlertDataSourceTesting{cfg: cfg}
}

func (dataSource *AlertDataSourceTesting) CoordinatorContractDataDB() (*coordinator.Storage, error) {
	if dataSource.cfg.CoordinatorContractDataDbFn == nil {
		return nil, errors.New("CoordinatorContractDataDbFn callback not set")
	}
	return dataSource.cfg.CoordinatorContractDataDbFn()
}

func (dataSource *AlertDataSourceTesting) FirstUnsignedPegoutDB() (*coordinator.PegoutRecord, error) {
	if dataSource.cfg.FirstUnsignedPegoutDbFn == nil {
		return nil, errors.New("FirstUnsignedPegoutDbFn callback not set")
	}
	return dataSource.cfg.FirstUnsignedPegoutDbFn()
}

func (dataSource *AlertDataSourceTesting) LastSignedPegoutDB() (*data_models.PegoutDbRow, error) {
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

func (dataSource *AlertDataSourceTesting) PegoutDB(address *address.Address) (*data_models.PegoutDbRow, error) {
	if dataSource.cfg.PegoutDbFn == nil {
		return nil, errors.New("PegoutDbFn callback not set")
	}
	return dataSource.cfg.PegoutDbFn(address)
}

func (dataSource *AlertDataSourceTesting) NowUnixTs() int64 {
	return dataSource.cfg.NowUnixTsFn()
}
