package fetchers

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	jwv4r2contract "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/jw_v4r2_contract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/config"
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
	/*
		jwv4r2contract, err := createJWV4R2Contract(tonClient, cfg.RelayerWalletV4Secret)
		var relayerAddr *address.Address
		if err != nil {
			logger.Log.Error().
				Msg(fmt.Sprintf("FetcherContractBalances: can't create jwv4r2contract, error: %v", err))
			relayerAddr = &address.Address{}
		} else {
			relayerAddr = jwv4r2contract.Address()
		}
	*/
	return &FetcherContractBalances{
		tonClient: tonClient,
		contractAddrs: map[string]*address.Address{
			"teleport":    cfg.TeleportContractAddr,
			"coordinator": cfg.CoordinatorContractAddr,
			"bitclient":   cfg.BitcoinClientContractAddr,
			"minter":      cfg.JettonMinterContractAddr,
			//"relayer":     relayerAddr,
		},
	}
}

func createJWV4R2Contract(tonClient *tonclient.TonClient, jwV4R2Secret string) (*jwv4r2contract.JWV4R2Contract, error) {
	jwV4R2SecretBytes, err := hex.DecodeString(jwV4R2Secret)
	if err != nil {
		return nil, fmt.Errorf("failed to decode jwv4r2 secret: %w", err)
	}

	jwV4R2Contract, err := jwv4r2contract.NewJWV4R2Contract(
		tonClient.API,
		jwV4R2SecretBytes,
		context.Background(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create jwv4r2 contract: %w", err)
	}
	return jwV4R2Contract, nil
}

func (fetcher *FetcherContractBalances) GetBalances() (map[string]float64, error) {
	balances := make(map[string]float64)

	for name, contractAddr := range fetcher.contractAddrs {
		balances[name] = -1.0 // Default value

		if contractAddr == nil {
			continue
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
			// TOOD: write to DB

			/*
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
			*/

			// TODO: reimplement with time.NewTicker
			time.Sleep(10 * time.Second)
		}
	}
}
