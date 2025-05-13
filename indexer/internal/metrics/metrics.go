package metrics

import (
	"context"
	"strconv"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/config"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/bitcoinclientcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
)

type Metrics struct {
	tonClient    *tonclient.TonClient
	bitcoinClient *bitcoin.Client
	bitcoinClientContract *bitcoinclientcontract.BitcoinClientContract
	contractAddr map[string]string
}

func New(
	tonClient *tonclient.TonClient,
	bitcoinClient *bitcoin.Client,
	bitclientContract *bitcoinclientcontract.BitcoinClientContract,
	config config.IndexerConfig,
) *Metrics {
	return &Metrics{
		tonClient: tonClient,
		bitcoinClient:    bitcoinClient,
		bitcoinClientContract: bitclientContract,
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
)

var (
	confirmationsNeededGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "confirmations_needed",
		Help: "Confirmations needed",
	})
)

var (
	lastConfirmedBlockHashGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "last_confirmed_block_hash",
		Help: "Last confirmed block hash",
	}, []string{"hash"})
)

var (
	candidateBlockHashesGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "candidate_block_hashes",
		Help: "Candidate block hashes",
	}, []string{"hash"})
)

var(
	lastConfirmedBlockHeightGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "last_confirmed_block_height",
		Help: "Last confirmed block height",
	})
)

var(
		chainName = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "chain_name",
			Help: "Chain Name",
		}, []string{"chain_name"})
)

var(
	chainHeight = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "chain_height",
		Help: "Chain Height",
	})
)

var(
	chainBestBlockHash = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "chain_best_block_hash",
		Help: "Chain Best Block Hash",
	},[]string{"hash"})
)

var(
	chainMedianTime = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "chain_median_time",
		Help: "Chain Median Time",
	})
)

func (m *Metrics) getBalances() (map[string]float64) {
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
	return balances
}

func (m *Metrics) recordBlockchainInfo(result *btcjson.GetBlockChainInfoResult) (error) {

	chainName.WithLabelValues(result.Chain).Set(1)
	chainHeight.Set(float64(result.Blocks))
	chainBestBlockHash.Reset()
	chainBestBlockHash.WithLabelValues(result.BestBlockHash).Set(1)
	chainMedianTime.Set(float64(result.MedianTime * 1000))

	return nil
}

func (m *Metrics) recordBalances(balances map[string]float64) (err error) {
	for key, value := range balances {
		contractBalances.WithLabelValues(utils.AddrToRawString(address.MustParseAddr(m.contractAddr[key])), key).Set(value)
	}
	return nil
}

func (m *Metrics) recordConfirmationsNeeded(confirmations int64) (err error) {
	confirmationsNeededGauge.Set(float64(confirmations))
	return nil
}

func (m *Metrics) recordLastConfirmedBlockHash(lastConfirmedBlockHash *chainhash.Hash) (err error) {
	lastConfirmedBlockHashGauge.Reset()
	lastConfirmedBlockHashGauge.WithLabelValues(lastConfirmedBlockHash.String()).Set(1)
	return nil
}

func (m *Metrics) recordCandidatedBlockHashes(candidates []*chainhash.Hash) (err error) {
	for _, value := range candidates {
		candidateBlockHashesGauge.WithLabelValues(value.String()).Set(1)
	}
	return nil
}

func (m *Metrics) recordLastConfirmedBlockHeight(lastConfirmedBlockHeight int64) (err error) {
	lastConfirmedBlockHeightGauge.Set(float64(lastConfirmedBlockHeight))
	return nil
}

func (m *Metrics) Work(ctx context.Context) (err error) {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			balances := m.getBalances()
			
			m.recordBalances(balances)
			confirmationsNeeded, err := m.bitcoinClientContract.GetConfirmationsNeeded()
			if err != nil {
				return m.formatGetConfirmationsNeededError()
			}
			m.recordConfirmationsNeeded(confirmationsNeeded)
			lastConfirmedBlockHash, err := m.bitcoinClientContract.GetLastConfirmedBlockHash()
			if err != nil {
				return m.formatGetLastConfirmedBlockHashError()
			}
			m.recordLastConfirmedBlockHash(lastConfirmedBlockHash)
			candidateBlockHashes, err := m.bitcoinClientContract.GetCandidateBlockHashes()
			if err != nil {
				return m.formatGetCandidateBlockHashesError()
			}
			m.recordCandidatedBlockHashes(candidateBlockHashes)
			lastConfirmedBlockHeight, err := m.bitcoinClient.GetBlockHeightByHash(lastConfirmedBlockHash)
			if err != nil {
				return m.formatGetLastConfirmedBlockHeightError()
			}
			m.recordLastConfirmedBlockHeight(lastConfirmedBlockHeight)
			blockchainInfo, err := m.bitcoinClient.RPCClient.GetBlockChainInfo()
			if err != nil {
				return m.formatGetBlockChainInfoError()
			}
			m.recordBlockchainInfo(blockchainInfo)
			time.Sleep(10 * time.Second)
		}
	}
}
