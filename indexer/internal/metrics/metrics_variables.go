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
)
