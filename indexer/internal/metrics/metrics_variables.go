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
	confirmedBlockMismatch = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "confirmed_block_mismatch",
			Help: "Confirmed block mismatch",
		},
		[]string{"contract_block", "network_block"},
	)
)
