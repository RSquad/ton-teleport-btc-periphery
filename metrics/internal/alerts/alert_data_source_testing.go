package alerts

import (
	"errors"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
)

type (
	FirstUnsignedPegoutFn     func() (*coordinator.PegoutRecord, error)
	CoordinatorContractDataFn func() (*coordinator.Storage, error)
	DkgFn                     func() (*coordinator.DKG, error)
	PrevDkgFn                 func() (*coordinator.DKG, error)
	NowUnixTsFn               func() int64
)

// Config holds optional function callbacks
type AlertDataSourceTestingConfig struct {
	FirstUnsignedPegoutFn     FirstUnsignedPegoutFn
	CoordinatorContractDataFn CoordinatorContractDataFn
	DkgFn                     DkgFn
	PrevDkgFn                 PrevDkgFn
	NowUnixTsFn               NowUnixTsFn
}

type AlertDataSourceTesting struct {
	cfg AlertDataSourceTestingConfig
}

func NewAlertDataSourceTesting(cfg AlertDataSourceTestingConfig) *AlertDataSourceTesting {
	return &AlertDataSourceTesting{cfg: cfg}
}

func (ds *AlertDataSourceTesting) FirstUnsignedPegout() (*coordinator.PegoutRecord, error) {
	if ds.cfg.FirstUnsignedPegoutFn == nil {
		return nil, errors.New("FirstUnsignedPegout callback not set")
	}
	return ds.cfg.FirstUnsignedPegoutFn()
}

func (ds *AlertDataSourceTesting) CoordinatorContractData() (*coordinator.Storage, error) {
	if ds.cfg.CoordinatorContractDataFn == nil {
		return nil, errors.New("CoordinatorContractData callback not set")
	}
	return ds.cfg.CoordinatorContractDataFn()
}

func (ds *AlertDataSourceTesting) Dkg() (*coordinator.DKG, error) {
	if ds.cfg.DkgFn == nil {
		return nil, errors.New("Dkg callback not set")
	}
	return ds.cfg.DkgFn()
}

func (ds *AlertDataSourceTesting) PrevDkg() (*coordinator.DKG, error) {
	if ds.cfg.PrevDkgFn == nil {
		return nil, errors.New("PrevDkg callback not set")
	}
	return ds.cfg.PrevDkgFn()
}

func (ds *AlertDataSourceTesting) NowUnixTs() int64 {
	return ds.cfg.NowUnixTsFn()
}
