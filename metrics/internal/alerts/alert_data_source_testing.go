package alerts

import (
	"errors"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
)

type (
	FirstUnsignedPegoutFn      func() (*coordinator.PegoutRecord, error)
	ConfiguratorContractDataFn func() (*coordinator.Storage, error)
	DkgFn                      func() (*coordinator.DKG, error)
	PrevDkgFn                  func() (*coordinator.DKG, error)
)

// Config holds optional function callbacks
type AlertDataSourceTestingConfig struct {
	FirstUnsignedPegoutFn      FirstUnsignedPegoutFn
	ConfiguratorContractDataFn ConfiguratorContractDataFn
	DkgFn                      DkgFn
	PrevDkgFn                  PrevDkgFn
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

func (ds *AlertDataSourceTesting) ConfiguratorContractData() (*coordinator.Storage, error) {
	if ds.cfg.ConfiguratorContractDataFn == nil {
		return nil, errors.New("ConfiguratorContractData callback not set")
	}
	return ds.cfg.ConfiguratorContractDataFn()
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
