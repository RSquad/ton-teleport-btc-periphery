package metrics

import (
	"context"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/config"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
)

type Metrics struct {
	tonClient        *tonclient.TonClient
	bitcoinClient    *bitcoin.Client
	teleportContract *teleportcontract.TeleportContract
	contractAddr     map[string]string
}

type txStatus struct {
	isCreated float64
	txID      string
}

func New(
	tonClient *tonclient.TonClient,
	bitcoinClient *bitcoin.Client,
	teleportContract *teleportcontract.TeleportContract,
	config config.IndexerConfig,
) *Metrics {
	return &Metrics{
		tonClient:        tonClient,
		bitcoinClient:    bitcoinClient,
		teleportContract: teleportContract,
		contractAddr: map[string]string{
			"teleport":    config.TeleportContractAddr,
			"coordinator": config.CoordinatorContractAddr,
			"bitclient":   config.BitcoinClientContractAddr,
			"minter":      config.JettonMinterContractAddr,
		},
	}
}

var (
	contractBalances = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "contract_balance",
		Help: "Contract balance",
	}, []string{"addr", "name"})

	pegoutCounter = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pegout_counter",
		Help: "Pegout counter",
	})

	txCreated = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "tx_created",
		Help: "Is transaction in mempool (1) or not (0)",
	}, []string{"tx_id"})
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

func (m *Metrics) getTxStatus(storage teleportcontract.Storage) (txStatus, error) {
	isTxFound := 0
	txVerbose, err := m.bitcoinClient.RPCClient.GetRawTransaction(storage.LastPegoutTxID)
	if err != nil {
		return txStatus{0, storage.LastPegoutTxID.String()},
			m.formatGetTxError(storage.LastPegoutTxID.String())
	}
	if txVerbose != nil {
		isTxFound = 1
	}

	return txStatus{float64(isTxFound), storage.LastPegoutTxID.String()}, nil
}

func (m *Metrics) recordMetrics(
	balances map[string]float64,
	pegouts float64,
	txStatus txStatus,
) (err error) {
	for key, value := range balances {
		contractBalances.WithLabelValues(
			utils.AddrToRawString(address.MustParseAddr(m.contractAddr[key])),
			key,
		).Set(value)
	}
	pegoutCounter.Set(pegouts)
	txCreated.WithLabelValues(txStatus.txID).Set(txStatus.isCreated)
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
			storage, err := m.teleportContract.GetStorage(nil)
			if err != nil {
				return err
			}
			txStatus, err := m.getTxStatus(storage)
			if err != nil {
				return err
			}
			m.recordMetrics(
				balances,
				float64(storage.PegoutChainCounter),
				txStatus,
			)
			time.Sleep(10 * time.Second)
		}
	}
}
