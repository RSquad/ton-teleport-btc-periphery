package alerts

import (
	"encoding/hex"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
	"github.com/xssnick/tonutils-go/address"
)

type AlertCpfpLength struct {
	lastUpdated time.Time
	severity    Severity
	labels      Labels
}

func NewAlertCpfpLength() Alert {
	return &AlertCpfpLength{
		lastUpdated: time.Time{},
		severity:    SEVERITY_OK,
		labels:      Labels{},
	}
}

func (alert *AlertCpfpLength) NewLabels() Labels {
	return Labels{
		"bitcoin_tx_id": "",
		"pegout_addr":   "",
	}
}

func (alert *AlertCpfpLength) Check(dataSource AlertDataSource) (Severity, Labels, Values, error) {
	labels := alert.NewLabels()

	pegout, err := dataSource.LastSignedPegoutDB()
	if err != nil {
		return SEVERITY_UNKNOWN, labels, nil, err
	}

	if pegout == nil {
		return SEVERITY_OK, labels, nil, nil
	}

	if pegout.BitcoinTxId == nil {
		return SEVERITY_OK, labels, nil, nil
	}

	if time.Since(alert.lastUpdated) < 2*time.Minute {
		return alert.severity, alert.labels, nil, nil
	}

	chainSize, err := dataSource.BtcGetCpfpLength(mutils.BytesToBTCHash(pegout.BitcoinTxId))

	if err != nil {
		return SEVERITY_UNKNOWN, labels, nil, err
	}

	if pegout.BitcoinTxId != nil {
		labels["bitcoin_tx_id"] = hex.EncodeToString(pegout.BitcoinTxId)
	}
	labels["pegout_addr"] = (*address.Address)(pegout.Addr).StringRaw()

	severity := alert.GetSeverity(chainSize)

	alert.lastUpdated = time.Now()
	alert.severity = severity
	alert.labels = labels

	return severity, labels, nil, nil
}

func (alert *AlertCpfpLength) GetSeverity(chainSize int) Severity {
	severity := SEVERITY_OK

	if chainSize >= 10 {
		severity = SEVERITY_WARNING
	}

	if chainSize >= 20 {
		severity = SEVERITY_CRITICAL
	}

	return severity
}
