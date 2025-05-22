package metrics

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/config"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
)

type FetcherContractBalances struct {
	tonClient     *tonclient.TonClient
	contractAddrs map[string]string
}

func NewFetcherContractBalances(tonClient *tonclient.TonClient, config config.IndexerConfig) *FetcherContractBalances {
	return &FetcherContractBalances{
		tonClient: tonClient,
		contractAddrs: map[string]string{
			"teleport":    config.TeleportContractAddr,
			"coordinator": config.CoordinatorContractAddr,
			"bitclient":   config.BitcoinClientContractAddr,
			"minter":      config.JettonMinterContractAddr,
		},
	}
}

func (fetcher *FetcherContractBalances) GetBalances() (map[string]float64, error) {
	balances := make(map[string]float64)

	for key, value := range fetcher.contractAddrs {
		balances[key] = -1.0 // Default value

		contractAddr := address.MustParseAddr(value)
		balance, err := fetcher.tonClient.GetBalance(contractAddr)
		if err != nil {
			logger.Log.Error().
				Msg(fmt.Sprintf("FetcherContractBalances: can`t get balance for contract: %s, error: %v", utils.AddrToRawString(contractAddr), err))
			continue
		}

		// TODO: make integral value (not float)
		balanceFloat, err := strconv.ParseFloat(balance.String(), 64)
		if err != nil {
			logger.Log.Error().Msg(fmt.Sprintf("FetcherContractBalances: can`t convert string: %s to float, error: %v", balance.String(), err))
			continue
		}

		balances[key] = balanceFloat
	}

	return balances, nil
}

func (fetcher *FetcherContractBalances) Work(ctx context.Context, wg *sync.WaitGroup) (err error) {
	defer wg.Done()

	defer logger.Log.Info().Msg("FetcherContractBalances: stopped")
	logger.DefaultLogStartWork("FetcherContractBalances: starting...")

	contractBalances := promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "contract_balance",
			Help: "Contract balance",
		},
		[]string{"addr", "name"},
	)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			balances, err := fetcher.GetBalances()
			if err != nil {
				return err
			}

			for key, value := range balances {
				contractBalances.WithLabelValues(utils.AddrToRawString(address.MustParseAddr(fetcher.contractAddrs[key])), key).Set(value)
			}

			// TODO: reimplement with time.NewTicker
			time.Sleep(10 * time.Second)
		}
	}
}
