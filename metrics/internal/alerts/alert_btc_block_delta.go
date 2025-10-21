package alerts

import "fmt"

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
		alert.severity = SEVERITY_CRITICAL
		alert.err = err

		return alert.severity, alert.description, alert.values, alert.err
	}

	blockHeightContract := int(storage.LastConfirmedBlockHeight + storage.ConfirmationsNeeded)
	blockHeightBtcNetwork := 0
	delta := 0
	for tryId := 1; tryId <= 5; tryId++ {
		var err error
		blockHeightBtcNetwork, err = dataSource.BtcGetBestBlockHeight()

		if err == nil {
			break
		}

		alert.severity = SEVERITY_CRITICAL
		alert.err = err

		if tryId == 5 {
			return alert.severity, alert.description, alert.values, alert.err
		}
	}

	delta = blockHeightBtcNetwork - blockHeightContract

	alert.severity = alert.GetSeverity(delta)

	if alert.severity > SEVERITY_OK {
		alert.description = Description(
			fmt.Sprintf(
				"There is a block-height delta of %d between the BitcoinClient contract (height %d: %d blocks + %d confirmations) and the Bitcoin network (height %d).",
				delta,
				blockHeightContract,
				storage.LastConfirmedBlockHeight,
				storage.ConfirmationsNeeded,
				blockHeightBtcNetwork,
			),
		)
	} else {

	}

	alert.err = nil

	return alert.severity, alert.description, alert.values, alert.err
}

func (alert *AlertBtcBlockDelta) GetSeverity(delta int) Severity {
	severity := SEVERITY_OK

	if delta == 2 {
		severity = SEVERITY_WARNING
	}

	if delta >= 3 {
		severity = SEVERITY_CRITICAL
	}

	return severity
}
