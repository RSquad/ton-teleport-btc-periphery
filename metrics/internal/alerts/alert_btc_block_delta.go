package alerts

type AlertBtcBlockDelta struct {
	lastUpdateTs int64
	severity     Severity
	labels       Labels
	vaules       Values
	err          error
}

func NewAlertBtcBlockDelta() Alert {
	a := AlertBtcBlockDelta{}

	return &AlertBtcBlockDelta{
		lastUpdateTs: 0,
		severity:     SEVERITY_UNKNOWN,
		labels:       a.NewLabels(),
		vaules:       nil,
		err:          nil,
	}
}

func (alert *AlertBtcBlockDelta) NewLabels() Labels {
	return Labels{
		"blockHash": "",
	}
}

func (alert *AlertBtcBlockDelta) Check(dataSource AlertDataSource) (Severity, Labels, Values, error) {
	nowTs := dataSource.NowUnixTs()

	// No more often than every 2 minutes
	if (nowTs - alert.lastUpdateTs) < (2 * 60) {
		return alert.severity, alert.labels, alert.vaules, alert.err
	}

	alert.lastUpdateTs = nowTs
	alert.labels = alert.NewLabels()
	alert.err = nil

	storage, err := dataSource.BitcoinClientContractStorageDB()

	if err != nil {
		alert.severity = SEVERITY_UNKNOWN
		alert.err = err

		return alert.severity, alert.labels, alert.vaules, alert.err
	}

	blockHeightContract := int(storage.LastConfirmedBlockHeight + storage.ConfirmationsNeeded)
	blockHeightNetwork, err := dataSource.BtcGetBestBlockHeight()

	if err != nil {
		alert.severity = SEVERITY_UNKNOWN
		alert.err = err

		return alert.severity, alert.labels, alert.vaules, alert.err
	}

	delta := blockHeightNetwork - blockHeightContract

	alert.severity = alert.GetSeverity(delta)
	alert.labels["blockHash"] = storage.LastConfirmedBlockHash.String()
	alert.err = nil

	return alert.severity, alert.labels, alert.vaules, alert.err
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
