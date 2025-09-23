package alerts

import (
	"encoding/hex"
	"sort"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
	"github.com/xssnick/tonutils-go/address"
)

type AlertPegoutInMempool struct {
	pegoutToCheck       *data_models.Pegout
	lastCheckedPegoutId uint64
	beginTimestamp      int64
}

func NewAlertPegoutInMempool() Alert {
	return &AlertPegoutInMempool{
		pegoutToCheck:       nil,
		lastCheckedPegoutId: 0,
		beginTimestamp:      0,
	}
}

func (alert *AlertPegoutInMempool) NewLabels() Labels {
	return Labels{
		"bitcoin_tx_id": "",
		"pegout_addr":   "",
	}
}

func (alert *AlertPegoutInMempool) Check(dataSource AlertDataSource) (Severity, Labels, Values, error) {
	labels := alert.NewLabels()

	if alert.pegoutToCheck == nil {
		// Get last signed pegouts
		pegouts, err := dataSource.LastSignedPegoutsDB(25)
		if err != nil {
			return SEVERITY_UNKNOWN, labels, nil, err
		}

		if len(pegouts) == 0 {
			return SEVERITY_OK, labels, nil, nil
		}

		// Sort by ID asc
		sort.Slice(pegouts, func(i, j int) bool {
			return pegouts[i].Id < pegouts[j].Id
		})

		// Get the earliest pegout
		for _, pegout := range pegouts {
			if alert.lastCheckedPegoutId >= pegout.Id {
				continue
			}

			if len(pegout.BitcoinTxId) == 0 {
				continue
			}

			alert.pegoutToCheck = pegout
			alert.lastCheckedPegoutId = pegout.Id
			alert.beginTimestamp = dataSource.NowUnixTs()
			break
		}

		if alert.pegoutToCheck == nil {
			return SEVERITY_OK, labels, nil, nil
		}
	}

	var isInMempoolOrBlock bool = false

	// Check mempool
	{
		btcMempoolEntry, err := dataSource.BtcGetMempoolEntry(mutils.BytesToBTCHash(alert.pegoutToCheck.BitcoinTxId).String())
		if err == nil {
			if btcMempoolEntry != nil {
				isInMempoolOrBlock = true
			}
		}
	}

	// Check block
	if !isInMempoolOrBlock {
		btcBlockHash, err := dataSource.BtcGetBlockHashByTxID(mutils.BytesToBTCHash(alert.pegoutToCheck.BitcoinTxId))
		if err == nil {
			if btcBlockHash != nil {
				isInMempoolOrBlock = true
			}
		}
	}

	// Update labels
	labels["bitcoin_tx_id"] = hex.EncodeToString(alert.pegoutToCheck.BitcoinTxId)
	labels["pegout_addr"] = (*address.Address)(alert.pegoutToCheck.Addr).StringRaw()

	// Calulate severity
	severity := SEVERITY_OK

	if !isInMempoolOrBlock {
		duration := time.Duration(dataSource.NowUnixTs()-alert.beginTimestamp) * time.Second
		severity = alert.GetSeverity(duration)
	} else {
		// isInMempoolOrBlock == true
		alert.pegoutToCheck = nil
	}

	return severity, labels, nil, nil
}

func (alert *AlertPegoutInMempool) GetSeverity(duration time.Duration) Severity {
	severity := SEVERITY_OK

	if duration >= 40*time.Minute {
		severity = SEVERITY_CRITICAL
	} else if duration >= 10*time.Minute {
		severity = SEVERITY_WARNING
	}

	return severity
}
