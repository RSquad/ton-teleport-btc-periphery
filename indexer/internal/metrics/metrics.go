package metrics

import (
	"context"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/config"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/xssnick/tonutils-go/address"
)

type Metrics struct {
	tonClient    *tonclient.TonClient
	contractAddr map[string]string
}

func New(tonClient *tonclient.TonClient, config config.IndexerConfig) *Metrics {
	return &Metrics{
		tonClient: tonClient,
		contractAddr: map[string]string{
			"teleport":    config.TeleportContractAddr,
			"coordinator": config.CoordinatorContractAddr,
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

	for key, value := range m.contractAddr {
		contractAddr := address.MustParseAddr(value)
		balance, err := m.tonClient.GetBalance(contractAddr)
		if err != nil {
			m.formatGetBalanceError(contractAddr)
		}

		balanceFloat, err := strconv.ParseFloat(balance.String(), 64)

		if err != nil {
			m.formatParseFloatError(balance.String())
		}

		balances[key] = balanceFloat
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
			m.recordMetrics(balances)
			time.Sleep(10 * time.Second)
		}
	}
}
