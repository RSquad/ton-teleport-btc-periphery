package prometheus

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/config"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/xssnick/tonutils-go/address"
)

type Metrics struct {
	tonClient         *tonclient.TonClient
	contractAddresses map[string]string
}

func New(tonClient *tonclient.TonClient, config config.IndexerConfig) *Metrics {
	return &Metrics{
		tonClient: tonClient,
		contractAddresses: map[string]string{
			"teleport":    config.TeleportContractAddr,
			"coordinator": config.CoordinatorContractAddr,
			"test":        "0QD3CxqLkN5V-jk24kdOlIIhNfGZYWH0y0ato9U_6pMBotZl",
		},
	}
}

var (
	contractBalances = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "contract_balance",
		Help: "Contract balance",
	}, []string{"contract"})
)

func (m *Metrics) getBalances() (map[string]float64, error) {
	balances := make(map[string]float64)

	for key, value := range m.contractAddresses {
		contractAddress, err := address.ParseAddr(value)
		if err != nil {
			continue
		}
		fmt.Println("contractAddress: ", contractAddress)
		balance, err := m.tonClient.GetBalance(contractAddress)
		fmt.Println("balance: ", balance)
		if err != nil {
			return balances, fmt.Errorf(getBalanceError, err)
		}

		balanceFloat, _ := new(big.Float).SetInt(balance).Float64()

		balances[key] = balanceFloat / 1000000000
	}
	return balances, nil
}

func (m *Metrics) recordMetrics(balances map[string]float64) (err error) {
	for key, value := range balances {
		contractBalances.With(prometheus.Labels{"contract": key}).Set(value)
	}
	return nil
}

func (m *Metrics) Work(ctx context.Context) (err error) {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			balances, err := m.getBalances()
			if err != nil {
				return err
			}
			fmt.Println("balances: ", balances["teleport"], " ", balances["coordinator"])
			m.recordMetrics(balances)
			time.Sleep(10 * time.Second)
		}
	}
}
