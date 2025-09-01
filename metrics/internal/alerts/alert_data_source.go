package alerts

import (
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
	"github.com/xssnick/tonutils-go/address"
)

type AlertDataSource interface {
	CoordinatorContractDataDB() (*coordinator.Storage, error)
	FirstUnsignedPegoutDB() (*coordinator.PegoutRecord, error)
	LastSignedPegoutDB() (*data_models.PegoutDbRow, error)
	PegoutDB(address *address.Address) (*data_models.PegoutDbRow, error)
	DkgDB() (*coordinator.DKG, error)
	PrevDkgDB() (*coordinator.DKG, error)

	NowUnixTs() int64
}
