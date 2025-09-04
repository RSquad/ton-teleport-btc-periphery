package fetchers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/config"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
	"github.com/xssnick/tonutils-go/address"
)

type FetcherContractBalances struct {
	chDB          chan PayloadDB
	tonClient     *tonclient.TonClient
	contractAddrs map[string]*address.Address
	period        int64 // Fetch period (in seconds)
}

func NewFetcherContractBalances(
	chDB chan PayloadDB,
	tonClient *tonclient.TonClient,
	cfg *config.ServicesConfig,
) *FetcherContractBalances {
	return &FetcherContractBalances{
		chDB:      chDB,
		tonClient: tonClient,
		contractAddrs: map[string]*address.Address{
			"coordinator": cfg.CoordinatorContractAddr,
			"teleport":    cfg.TeleportContractAddr,
			"bitclient":   cfg.BitcoinClientContractAddr,
			"minter":      cfg.JettonMinterContractAddr,
			"relayer":     cfg.RelayerWalletAddr,
		},
		period: int64(cfg.ContractBalancesFetchPeriod),
	}
}

func (fetcher *FetcherContractBalances) Work(ctx context.Context, wg *sync.WaitGroup) (err error) {
	defer wg.Done()

	defer logger.Log.Info().Msg("FetcherContractBalances: stopped")
	logger.DefaultLogStartWork("FetcherContractBalances: starting...")

	ticker := time.NewTicker(time.Duration(fetcher.period) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info().Msg("FetcherContractBalances received shutdown signal...")
			return
		case <-ticker.C:
			fetcher.Fetch()
		}
	}
}

func (fetcher *FetcherContractBalances) Fetch() {
	balances, err := fetcher.GetBalances()
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherContractBalances").
			Msg("failed to GetBalances")
	}

	jsonData, err := data_models.SerializeContractBalancesDB(balances)
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherContractBalances").
			Msg("failed to serialize json")
	}

	fetcher.chDB <- PayloadDB{
		timestamp: time.Now(),
		typeId:    PayloadTypeContractBalances,
		payload:   string(jsonData),
	}
}

func (fetcher *FetcherContractBalances) GetBalances() (*data_models.ContractBalances, error) {
	var balances data_models.ContractBalances

	for name, contractAddr := range fetcher.contractAddrs {
		contractBalance := data_models.ContractBalance{
			Name:    name,
			Addr:    contractAddr,
			Balance: 0,
		}

		account, err := fetcher.tonClient.FetchAcc(contractAddr, nil)
		if err != nil {
			logger.Log.Error().
				Msg(fmt.Sprintf("FetcherContractBalances: can't fetch account %s, error: %v", utils.AddrToRawString(contractAddr), err))
			continue
		}

		if !account.IsActive {
			logger.Log.Error().
				Msg(fmt.Sprintf("FetcherContractBalances: account is not active %s, error: %v", utils.AddrToRawString(contractAddr), err))
			continue
		}

		balanceCoins, err := fetcher.tonClient.GetBalance(contractAddr)
		if err != nil {
			logger.Log.Error().
				Msg(fmt.Sprintf("FetcherContractBalances: can`t get balance for contract: %s, error: %v", utils.AddrToRawString(contractAddr), err))
			continue
		}

		contractBalance.Balance = balanceCoins.Nano().Uint64()

		balances.Balances = append(balances.Balances, &contractBalance)
	}

	return &balances, nil
}
