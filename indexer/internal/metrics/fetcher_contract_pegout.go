package metrics

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/pegoutcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/xssnick/tonutils-go/address"
)

type FetcherContractPegout struct {
	db        *sql.DB
	tonClient *tonclient.TonClient
	period    int64 // Fetch period (in seconds)
}

type PegoutData struct {
	Address string
}

func NewFetcherContractPegout(
	db *sql.DB,
	tonClient *tonclient.TonClient,
	period int64,
) *FetcherContractPegout {
	return &FetcherContractPegout{
		db:        db,
		tonClient: tonClient,
		period:    period,
	}
}

func (fetcher *FetcherContractPegout) Work(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	defer logger.Log.Info().Msg("FetcherContractPegout: stopped")
	logger.DefaultLogStartWork("FetcherContractPegout: starting...")

	ticker := time.NewTicker(time.Duration(fetcher.period) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info().Msg("FetcherContractPegout received shutdown signal...")
			return
		case <-ticker.C:
			fetcher.Fetch()
		}
	}
}

func (fetcher *FetcherContractPegout) createPegoutContract(pegoutAddr *address.Address) *pegoutcontract.PegoutContract {
	pegoutContract := pegoutcontract.New(
		pegoutAddr,
		fetcher.tonClient,
		context.Background(),
	)
	return pegoutContract
}

func (fetcher *FetcherContractPegout) getPegoutAddr() (*address.Address, error) {
	// rows, err := fetcher.db.Query(``) // TODO: create query to get pegout contract address

	// if err != nil {
	// 	return &address.Address{}, err
	// }

	// defer rows.Close()

	// var data string
	// if rows.Next() {
	// 	err = rows.Scan(&data)
	// 	if err != nil {
	// 		return &address.Address{}, err
	// 	}
	// }

	// var pegoutData PegoutData
	// err = json.Unmarshal([]byte(data), &pegoutData)
	// if err != nil {
	// 	return &address.Address{}, err
	// }
	// TODO: replace with pegoutData.Address
	pegoutAddr, err := address.ParseRawAddr("0:0cbe248cbeb717298a4dd52e98d11904fe93817c64b594a3f98a283a6177e4a0")
	if err != nil {
		logger.Log.Error().Msg(fmt.Sprintf("FetcherContractPegout: failed to parse pegout address, error: %v", err))
		return &address.Address{}, nil
	}
	return pegoutAddr, nil
}

func (fetcher *FetcherContractPegout) Fetch() {
	pegoutAddr, err := fetcher.getPegoutAddr()
	if err != nil {
		logger.Log.Error().Msg(fmt.Sprintf("FetcherContractPegout: failed to retrieve pegout address, error: %v", err))
		return
	}
	pegoutContract := fetcher.createPegoutContract(pegoutAddr)
	storage, err := pegoutContract.GetStorage(nil)
	if err != nil {
		logger.Log.Error().Msg(fmt.Sprintf("FetcherContractPegout: failed to retrieve storage cell, error: %v", err))
		return
	}

	txParts, err := pegoutContract.GetTxParts(nil)
	if err != nil {
		logger.Log.Error().Msg(fmt.Sprintf("FetcherContractPegout: failed to retrieve tx parts, error: %v", err))
		return
	}

	totalInputAmount := new(big.Int)
	inputs, err := txParts.Inputs.ToSortedSlice()
	if err != nil {
		logger.Log.Error().Msg(fmt.Sprintf("FetcherContractPegout: failed to sort inputs, error: %v", err))
		return
	}

	for _, input := range inputs {
		totalInputAmount.Add(totalInputAmount, input.Data.Amount)
	}

	totalOutputAmount := new(big.Int)
	for _, output := range txParts.Outputs {
		totalOutputAmount.Add(totalOutputAmount, big.NewInt(int64(output.Amount)))
	}

	pegoutInputsOutputsMismatch.
		WithLabelValues(pegoutContract.Addr.String()).
		Set(float64(totalInputAmount.Cmp(totalOutputAmount.Add(totalOutputAmount, storage.TxFee))))
}
