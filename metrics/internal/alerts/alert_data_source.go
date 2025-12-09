package alerts

import (
	"context"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
	"github.com/xssnick/tonutils-go/address"
)

type AlertDataSource interface {
	CoordinatorContractStorageDB() (*coordinator.Storage, error)
	TeleportContractStorageDB() (*teleportcontract.Storage, error)
	BitcoinClientContractStorageDB() (*data_models.BitcoinClientContractStorage, error)
	FirstUnsignedPegoutDB() (*coordinator.PegoutRecord, error)
	LastConfirmedPegout() (*data_models.Pegout, error)
	LastSignedPegoutDB() (*data_models.Pegout, error)
	LastSignedPegoutsDB(limit uint) ([]*data_models.Pegout, error)
	PegoutDB(address *address.Address) (*data_models.Pegout, error)
	DkgDB() (*coordinator.DKG, error)
	DkgUntilDB(dkgUntil time.Time) (*coordinator.DKG, error)
	PrevDkgDB() (*coordinator.DKG, error)
	BtcGetBlockHashByTxID(txID *chainhash.Hash) (*chainhash.Hash, error)
	BtcGetBlockHeightByHash(hash *chainhash.Hash) (int64, error)
	BtcGetBestBlockHeight() (int, error)
	BtcGetCpfpLength(hash *chainhash.Hash) (int, error)
	BtcGetMempoolEntry(txHash string) (*btcjson.GetMempoolEntryResult, error)
	TonMaxMainValidators(ctx context.Context) (int, error)
	ActualContractBalance(name string) (int64, error)

	NowUnixTs() int64

	// Events
	EventsLastDkgStartedDB() (*coordinator.DKGStartedEvent, error)
	EventsAllFromDkgRestartDB(fromTxLT uint64) ([]*coordinator.DKGRestartedEvent, error)
}
