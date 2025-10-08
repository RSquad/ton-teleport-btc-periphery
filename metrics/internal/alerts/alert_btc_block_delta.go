package alerts

import "math"

type AlertBtcBlockDelta struct {
}

func NewAlertBtcBlockDelta() Alert {
	return &AlertBtcBlockDelta{}
}

func (alert *AlertBtcBlockDelta) NewLabels() Labels {
	return Labels{
		"blockHash": "",
	}
}

func (alert *AlertBtcBlockDelta) Check(dataSource AlertDataSource) (Severity, Labels, Values, error) {
	labels := alert.NewLabels()

	delta := 0
	blockHeightContract := 0
	blockHeightNetwork := 0
	blockHash := ""

	storage, err := dataSource.BitcoinClientContractStorageDB()

	if err != nil {
		return SEVERITY_UNKNOWN, labels, nil, err
	}

	blockHeightContract = int(storage.LastConfirmedBlockHeight)
	blockHeightNetwork, err = dataSource.BtcGetBestBlockHeight()

	for i := range storage.CandidateBlockHashes {
		canditateHeight, err := dataSource.BtcGetBlockHeightByHash(storage.CandidateBlockHashes[i])
		if err != nil {
			return SEVERITY_UNKNOWN, labels, nil, err
		}
		if int(canditateHeight) > blockHeightNetwork {
			blockHeightNetwork = int(canditateHeight)
			blockHash = storage.CandidateBlockHashes[i].String()
		}
	}

	if err != nil {
		return SEVERITY_UNKNOWN, labels, nil, err
	}

	delta = blockHeightNetwork - blockHeightContract

	severity := alert.GetSeverity(int(math.Abs(float64(delta))))
	labels["blockHash"] = blockHash

	return severity, labels, nil, nil
}

func (alert *AlertBtcBlockDelta) GetSeverity(delta int) Severity {
	severity := SEVERITY_OK

	if delta == 1 {
		severity = SEVERITY_WARNING
	}

	if delta >= 2 {
		severity = SEVERITY_CRITICAL
	}

	return severity
}
