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

func recordMetrics(balances map[string]float64) {
	go func() {
			for key, value := range balances {
					contractBalances.With(prometheus.Labels{"contract_name": key}).Set(value)
			}
	}()
}

var (
	contractBalances = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "teleport_contracts_balances",
			Help: "Current contracts balances",
	}, []string{"contract_name"})
)

func (p *Prometheus) Work(ctx context.Context) (err error) {
	v := reflect.ValueOf(p.config)
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
		
		balance, err := p.getContractBalance(ctx, contractAddress)
		if err != nil {
			return fmt.Errorf("failed to get contract balance: %v", err)
		}
		balances[fieldName] = float64(balance) / 1000000000
		fmt.Println("balances: ", balances)
		recordMetrics(balances)
	
	}
	return nil
}

func (p *Prometheus) getContractBalance(ctx context.Context, addr *address.Address) (int64, error) {
	block, err := p.tonClient.API.CurrentMasterchainInfo(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to get masterchain info: %v", err)
	}
	
	account, err := p.tonClient.API.GetAccount(ctx, block, addr)
	if err != nil {
		return 0, fmt.Errorf("failed to get account: %v", err)
	}

	return account.State.Balance.Nano().Int64(), nil
}
