package prometheus

import (
	"context"
	"fmt"
	"reflect"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/config"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/xssnick/tonutils-go/address"
)

type Prometheus struct {
	tonClient *tonclient.TonClient
	config config.IndexerConfig
}

func New(tonClient *tonclient.TonClient, config config.IndexerConfig) *Prometheus {
	return &Prometheus{
		tonClient: tonClient,
		config: config,
	}
}

func recordMetrics(balance int64) {
	go func() {
			for {
					contractBalances.Set(float64(balance))
			}
	}()
}

var (
	contractBalances = promauto.NewGauge(prometheus.GaugeOpts{
			Name: "teleport_contracts_balances",
			Help: "Current contract balances",
	})
)

func (p *Prometheus) Work(ctx context.Context) (err error) {
	v := reflect.ValueOf(p.config)

	for i := range v.NumField() {
		field := v.Field(i)
		var fieldValue string = field.Interface().(string)
		contractAddress, err := address.ParseAddr(fieldValue)

		if err != nil {
			continue
		}
		
		balance, err := p.getContractBalance(ctx, contractAddress)
		recordMetrics(balance)
		if err != nil {
			return fmt.Errorf("failed to record metrics: %v", err)
		}
	
	}
	return nil
}

func (p *Prometheus) getContractBalance(ctx context.Context, addr *address.Address) (int64, error) {
	// Get the account state
	block, err := p.tonClient.API.CurrentMasterchainInfo(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to get masterchain info: %v", err)
	}
	
	account, err := p.tonClient.API.GetAccount(ctx, block, addr)
	if err != nil {
		return 0, fmt.Errorf("failed to get account: %v", err)
	}

	// Return the balance in nanoTON
	return account.State.Balance.Nano().Int64(), nil
}
