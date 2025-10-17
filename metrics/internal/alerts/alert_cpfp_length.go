package alerts

import (
	"encoding/hex"
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
	"github.com/xssnick/tonutils-go/address"
)

type AlertCpfpLength struct {
	lastUpdateTs int64
	severity     Severity
	labels       Labels
	values       Values
	err          error
}

func NewAlertCpfpLength() Alert {
	a := AlertCpfpLength{}

	return &AlertCpfpLength{
		lastUpdateTs: 0,
		severity:     SEVERITY_UNKNOWN,
		labels:       a.NewLabels(),
		values:       nil,
		err:          nil,
	}
}

func (alert *AlertCpfpLength) NewLabels() Labels {
	return Labels{
		"bitcoin_tx_id": "",
		"pegout_addr":   "",
	}
}

func (alert *AlertCpfpLength) Check(dataSource AlertDataSource) (Severity, Labels, Values, error) {
	nowTs := dataSource.NowUnixTs()

	// Not more often than once every 2 minutes
	if (nowTs - alert.lastUpdateTs) < (2 * 60) {
		return alert.severity, alert.labels, alert.values, alert.err
	}

	alert.lastUpdateTs = nowTs
	alert.labels = alert.NewLabels()
	alert.err = nil
	alert.values = nil

	pegout, err := dataSource.LastConfirmedPegout()
	if err != nil {
		alert.severity = SEVERITY_UNKNOWN
		alert.err = err
		return alert.severity, alert.labels, alert.values, alert.err
	}

	if pegout == nil {
		alert.severity = SEVERITY_OK
		alert.err = nil
		return alert.severity, alert.labels, alert.values, alert.err
	}

	alert.labels["pegout_addr"] = (*address.Address)(pegout.Addr).StringRaw()

	if pegout.BitcoinTxId == nil {
		alert.severity = SEVERITY_UNKNOWN
		alert.err = fmt.Errorf("bitcoin TxId is null")
		return alert.severity, alert.labels, alert.values, alert.err
	}

	alert.labels["bitcoin_tx_id"] = hex.EncodeToString(pegout.BitcoinTxId)

	chainSize, err := dataSource.BtcGetCpfpLength(mutils.BytesToBTCHash(pegout.BitcoinTxId))

	if err != nil {
		alert.severity = SEVERITY_UNKNOWN
		alert.err = err
		return alert.severity, alert.labels, alert.values, alert.err
	}

	alert.severity = alert.GetSeverity(chainSize)
	alert.err = nil

	return alert.severity, alert.labels, alert.values, alert.err
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
