package alerts

import (
	"encoding/hex"
	"fmt"
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

func (alert *AlertPegoutInMempool) Check(dataSource AlertDataSource) (Severity, Description, Values, error) {
	if alert.pegoutToCheck == nil {
		// Get last signed pegouts
		pegouts, err := dataSource.LastSignedPegoutsDB(25)
		if err != nil {
			return SEVERITY_CRITICAL, "", nil, err
		}

		if len(pegouts) == 0 {
			return SEVERITY_OK, "OK", nil, nil
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
			return SEVERITY_OK, "OK", nil, nil
		}
	}

	var isInMempoolOrBlock bool = false

	// Check alert.pegoutToCheck.BitcoinTxId
	if len(alert.pegoutToCheck.BitcoinTxId) != 32 {
		return SEVERITY_CRITICAL, "", nil, fmt.Errorf("wrong BitcoinTxId value '%s'", hex.EncodeToString(alert.pegoutToCheck.BitcoinTxId))
	}

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

	// Calulate severity
	severity := SEVERITY_OK

	if !isInMempoolOrBlock {
		duration := time.Duration(dataSource.NowUnixTs()-alert.beginTimestamp) * time.Second
		severity = alert.GetSeverity(duration)
	} else {
		alert.pegoutToCheck = nil
	}

	description := "OK"

	if severity > SEVERITY_OK {
		timeout := 10
		if severity == SEVERITY_CRITICAL {
			timeout = 40
		}

		description = fmt.Sprintf(
			"Pegout transaction has not been found in the mempool for more than %d minutes.\nPegout: %s.\nBitcoin TX: %s.\nRunbook url: %s",
			timeout,
			mutils.TonExplorerLink((*address.Address)(alert.pegoutToCheck.Addr).StringRaw()),
			mutils.BtcExplorerLink(hex.EncodeToString(alert.pegoutToCheck.BitcoinTxId)),
			mutils.RunbookLink("PegoutInMempool"),
		)
	}

	return severity, Description(description), nil, nil
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
