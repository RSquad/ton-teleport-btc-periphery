package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	dkgStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dkg_status",
			Help: "DKG Status",
		},
		[]string{"dkg_status"},
	)
	contractBalances = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "contract_balance",
			Help: "Contract balance",
		},
		[]string{"addr", "name"},
	)
	unsignedPegoutsLen = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pegouts_len",
			Help: "Pegouts len",
		},
		[]string{"len"},
	)
	unsignedPegoutDelayed = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pegout_delayed",
			Help: "Pegout delayed",
		},
		[]string{"pegout_addr"},
	)
	unprocessedPegout = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "unprocessed_pegout",
			Help: "Unprocessed pegout",
		},
		[]string{"pegout_addr", "bitcoin_tx_id"},
	)
	lastBlockHeightDifference = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "last_block_height_difference",
			Help: "Last block height difference",
		},
	)
	utxoKeysDifference = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "utxo_keys_difference",
			Help: "Utxo keys difference",
		},
		[]string{"expected", "found"},
	)
)
