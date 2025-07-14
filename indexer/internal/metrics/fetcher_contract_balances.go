package metrics

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/config"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
)

type FetcherContractBalances struct {
	tonClient     *tonclient.TonClient
	contractAddrs map[string]*address.Address
}

func NewFetcherContractBalances(
	tonClient *tonclient.TonClient,
	cfg *config.ServicesConfig,
) *FetcherContractBalances {
	return &FetcherContractBalances{
		tonClient: tonClient,
		contractAddrs: map[string]*address.Address{
			"teleport":    cfg.ExternalServices.TeleportContractAddr,
			"coordinator": cfg.ExternalServices.CoordinatorContractAddr,
			"bitclient":   cfg.ExternalServices.BitcoinClientContractAddr,
			"minter":      cfg.ExternalServices.JettonMinterContractAddr,
		},
	}
}

func (fetcher *FetcherContractBalances) GetBalances() (map[string]float64, error) {
	balances := make(map[string]float64)

	for name, contractAddr := range fetcher.contractAddrs {
		balances[name] = -1.0 // Default value

		if contractAddr == nil {
			continue
		}

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

		balances[name] = balanceFloat
	}

	return balances, nil
}

func (fetcher *FetcherContractBalances) Work(ctx context.Context, wg *sync.WaitGroup) (err error) {
	defer wg.Done()

	defer logger.Log.Info().Msg("FetcherContractBalances: stopped")
	logger.DefaultLogStartWork("FetcherContractBalances: starting...")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			balances, err := fetcher.GetBalances()
			if err != nil {
				return err
			}

			for name, value := range balances {
				contractAddr := fetcher.contractAddrs[name]

				if contractAddr != nil {
					contractBalances.WithLabelValues(utils.AddrToRawString(contractAddr), name).Set(value)
				}
			}

			// TODO: reimplement with time.NewTicker
			time.Sleep(10 * time.Second)
		}
	}
}
