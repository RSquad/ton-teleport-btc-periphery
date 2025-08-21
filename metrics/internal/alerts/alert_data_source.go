package alerts

import "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"

type AlertDataSource interface {
	ConfiguratorContractData() (*coordinator.Storage, error)
	FirstUnsignedPegout() (*coordinator.PegoutRecord, error)
	PrevDkg() (*coordinator.DKG, error)
}
