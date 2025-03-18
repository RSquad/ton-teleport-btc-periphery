package prometheus

import (
	"context"
	"fmt"
	"math/big"
	"reflect"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/config"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/xssnick/tonutils-go/address"
)

type Metrics struct {
	tonClient *tonclient.TonClient
	config    config.IndexerConfig
}

func New(tonClient *tonclient.TonClient, config config.IndexerConfig) *Metrics {
	return &Metrics{
		tonClient: tonClient,
		config:    config,
	}
}

var (
	contractBalances = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "teleport_contracts_balances",
		Help: "Current contracts balances",
	}, []string{"contract_name"})
)

func (m *Metrics) fetchDynamicValue() (map[string]float64, error) {
	v := reflect.ValueOf(m.config)
	typeOfC := v.Type()
	balances := make(map[string]float64)

	for i := range v.NumField() {
		field := v.Field(i)
		var fieldValue string = field.Interface().(string)
		fieldName := typeOfC.Field(i).Name
		contractAddress, err := address.ParseAddr(fieldValue)
		if err != nil {
			continue
		}

		balance, err := m.tonClient.GetBalance(contractAddress)
		if err != nil {
			return balances, fmt.Errorf("failed to get contract balance: %v", err)
		}

		balanceFloat, _ := new(big.Float).SetInt(balance).Float64()

		balances[fieldName] = balanceFloat / 1000000000
	}
	return balances, nil
}

func (m *Metrics) recordMetrics() (err error) {
	balances, err := m.fetchDynamicValue()
	if err != nil {
		return err
	}
	go func() {
		for key, value := range balances {
			contractBalances.With(prometheus.Labels{"contract_name": key}).Set(value)
		}
	}()
	return nil
}

func (m *Metrics) Work(ctx context.Context) (err error) {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			m.recordMetrics()
			time.Sleep(10 * time.Second)
		}
	}
}
