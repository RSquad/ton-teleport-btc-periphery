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
	description  Description
	values       Values
	err          error
}

func NewAlertCpfpLength() Alert {
	return &AlertCpfpLength{
		lastUpdateTs: 0,
		severity:     SEVERITY_UNKNOWN,
		description:  "",
		values:       nil,
		err:          nil,
	}
}

func (alert *AlertCpfpLength) Check(dataSource AlertDataSource) (Severity, Description, Values, error) {
	nowTs := dataSource.NowUnixTs()
	component := "AlertCpfpLength"

	if alert.err == nil {
		// Not more often than once every 2 minutes
		if (nowTs - alert.lastUpdateTs) < (2 * 60) {
			return alert.severity, alert.description, alert.values, alert.err
		}
	}

	alert.lastUpdateTs = nowTs
	alert.description = "OK"
	alert.err = nil
	alert.values = nil

	pegout, err := dataSource.LastConfirmedPegout()
	if err != nil {
		alert.severity = SEVERITY_CRITICAL
		alert.err = err

		logPegoutFetchError(component, err)
		return alert.severity, "", alert.values, alert.err
	}

	if pegout == nil {
		alert.severity = SEVERITY_OK
		alert.err = nil

		logNoPegoutsFound(component)
		return alert.severity, alert.description, alert.values, alert.err
	}

	if pegout.BitcoinTxId == nil {
		alert.severity = SEVERITY_CRITICAL
		alert.err = fmt.Errorf("bitcoin TxId is null")

		logNullBitcoinTxId(pegout)
		return alert.severity, "", alert.values, alert.err
	}

	btcTxIdHex := hex.EncodeToString(pegout.BitcoinTxId)
	chainSize, err := dataSource.BtcGetCpfpLength(mutils.BytesToBTCHash(pegout.BitcoinTxId))
	if err != nil {
		alert.severity = SEVERITY_CRITICAL
		alert.err = err

		logCpfpLengthFetchError(btcTxIdHex, err)
		return alert.severity, "", alert.values, alert.err
	}

	alert.severity = alert.GetSeverity(chainSize)
	alert.err = nil
	alert.description = "OK"

	if alert.severity > SEVERITY_OK {
		alert.description = Description(fmt.Sprintf(
			"The CPFP chain length is %d.\n<b>Pegout:</b> %s.\n<b>Bitcoin TX:</b> %s.\n<b>Runbook url:</b> %s",
			chainSize,
			mutils.TonExplorerLink((*address.Address)(pegout.Addr).StringRaw()),
			mutils.BtcExplorerLink(hex.EncodeToString(pegout.BitcoinTxId)),
			mutils.RunbookLink("PegoutCPFP"),
		))
	}

	return alert.severity, alert.description, alert.values, alert.err
}

func (alert *AlertCpfpLength) GetSeverity(chainSize int) Severity {
	severity := SEVERITY_OK

	if chainSize >= 20 {
		severity = SEVERITY_CRITICAL
	} else if chainSize >= 10 {
		severity = SEVERITY_WARNING
	}

	return severity
}
