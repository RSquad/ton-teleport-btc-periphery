package alerts

import (
	"fmt"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
)

type AlertBtcBlockDelta struct {
	lastUpdateTs int64
	severity     Severity
	description  Description
	values       Values
	err          error
}

func NewAlertBtcBlockDelta() Alert {
	return &AlertBtcBlockDelta{
		lastUpdateTs: 0,
		severity:     SEVERITY_UNKNOWN,
		description:  "",
		values:       nil,
		err:          nil,
	}
}

func (alert *AlertBtcBlockDelta) Check(dataSource AlertDataSource) (Severity, Description, Values, error) {
	nowTs := dataSource.NowUnixTs()

	// No more often than every 2 minutes
	if alert.err == nil {
		if (nowTs - alert.lastUpdateTs) < (2 * 60) {
			return alert.severity, alert.description, alert.values, alert.err
		}
	}

	alert.lastUpdateTs = nowTs
	alert.description = "OK"
	alert.err = nil
	alert.values = nil

	storage, err := dataSource.BitcoinClientContractStorageDB()
	if err != nil {
		alert.description = ""
		alert.severity = SEVERITY_CRITICAL
		alert.err = err

		return alert.severity, alert.description, alert.values, alert.err
	}

	blockHeightContract := int(storage.LastConfirmedBlockHeight + storage.ConfirmationsNeeded)
	blockHeightBtcNetwork, err := alert.btcGetBestBlockHeight(dataSource)
	if err != nil {
		alert.description = ""
		alert.severity = SEVERITY_CRITICAL
		alert.err = err

		return alert.severity, alert.description, alert.values, alert.err
	}

	delta := blockHeightBtcNetwork - blockHeightContract
	alert.severity = alert.GetSeverity(delta)
	alert.err = nil
	alert.description = "OK"

	if alert.severity > SEVERITY_OK {
		alert.description = Description(
			fmt.Sprintf(
				"There is a block-height delta of %d between the BitcoinClient contract (height %d: %d blocks + %d confirmations) and the Bitcoin network (height %d).\nRunbook url: %s",
				delta,
				blockHeightContract,
				storage.LastConfirmedBlockHeight,
				storage.ConfirmationsNeeded,
				blockHeightBtcNetwork,
				mutils.RunbookLink("BtcBlockDelta"),
			),
		)
	}

	return alert.severity, alert.description, alert.values, alert.err
}

func (alert *AlertBtcBlockDelta) GetSeverity(delta int) Severity {
	severity := SEVERITY_OK

	if delta >= 3 {
		severity = SEVERITY_CRITICAL
	} else if delta >= 2 {
		severity = SEVERITY_WARNING
	}

	return severity
}

func (alert *AlertBtcBlockDelta) btcGetBestBlockHeight(dataSource AlertDataSource) (int, error) {
	var err error = nil
	blockHeightBtcNetwork := 0

	for tryId := 1; tryId <= 5; tryId++ {
		blockHeightBtcNetwork, err = dataSource.BtcGetBestBlockHeight()

		if err == nil {
			return blockHeightBtcNetwork, nil
		}

		time.Sleep(2 * time.Second)
	}

	return 0, err
}
