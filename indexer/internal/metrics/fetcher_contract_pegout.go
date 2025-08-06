package metrics

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/pegoutcontract"
)

type FetcherContractPegout struct {
	pegoutContract pegoutcontract.PegoutContract
	period         int64 // Fetch period (in seconds)
}

func NewFetcherContractPegout(
	pegoutContract pegoutcontract.PegoutContract,
	period int64,
) *FetcherContractPegout {
	return &FetcherContractPegout{
		pegoutContract: pegoutContract,
		period:         period,
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

func (fetcher *FetcherContractPegout) Fetch() {
	storage, err := fetcher.pegoutContract.GetStorage(nil)
	if err != nil {
		logger.Log.Error().Msg(fmt.Sprintf("FetcherContractPegout: failed to retrieve storage cell, error: %v", err))
		return
	}

	txParts, err := fetcher.pegoutContract.GetTxParts(nil)
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
		WithLabelValues(fetcher.pegoutContract.Addr.String()).
		Set(float64(totalInputAmount.Cmp(totalOutputAmount.Add(totalOutputAmount, storage.TxFee))))
}
