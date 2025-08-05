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
	dkgMaxSigners = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "dkg_signers_count",
			Help: "DKG signers count",
		},
	)
	totalValidatorsCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "total_validators_count",
			Help: "Total validators count",
		},
	)
	cpfpCounter = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cpfp_count",
			Help: "CPFP Count",
		},
		[]string{"tx_id"},
	)
	svbExceeded = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "svb_exceeded",
			Help: "SVB Exceeded",
		},
	)
	nextSvbNotZero = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "next_svb_not_zero",
			Help: "Next SVB not zero",
		},
		[]string{"bitcoin_tx_id"},
	)
	autopegoutDelayed = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "autopegout_delayed",
			Help: "Autopegout delayed",
		},
	)
	wrongInternalKey = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "wrong_internal_key",
			Help: "Wrong internal key",
		},
		[]string{"internal_key", "pegout_addr"},
	)
	insufficientValidators = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "insufficient_validators",
			Help: "Insufficient validators",
		},
	)
	pegoutSigningRestart = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pegout_signing_restart",
			Help: "Pegout signing restart",
		},
		[]string{"pegout_addr"},
	)
	pegoutRestartCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pegout_restart_count",
			Help: "Pegout restart count",
		},
		[]string{"pegout_addr"},
	)
	pegoutSigningMaskValidatorsCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pegout_signing_validators_count",
			Help: "Pegout signing validators count",
		},
		[]string{"pegout_addr"},
	)
	pegoutSigningCulpritGotThreshold = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pegout_signing_culprit",
			Help: "Pegout signing culprit",
		},
		[]string{"pegout_addr"},
	)
	pegoutSigningCulpritNotGetThreshold = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pegout_signing_culprit_not_get_threshold",
			Help: "Pegout signing culprit not get threshold",
		},
		[]string{"pegout_addr"},
	)
	dkgCulpritRemains = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "dkg_culprit remains in vset",
			Help: "DKG Culprit remains in vset",
		},
	)
	dkgCulpritRemoved = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "dkg_culprit removed from vset",
			Help: "DKG Culprit removed from vset",
		},
	)
)
