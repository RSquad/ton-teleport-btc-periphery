package alerts

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
)

type AlertPegoutRestarts struct {
	currentUnsignedPegout *coordinator.PegoutRecord
	restartsCounter       int
}

func NewAlertPegoutRestarts() Alert {
	return &AlertPegoutRestarts{
		currentUnsignedPegout: nil,
		restartsCounter:       0,
	}
}

func (alert *AlertPegoutRestarts) Check(dataSource AlertDataSource) (Severity, Description, Values, error) {
	// Get first unsigned pegout
	unsignedPegout, err := dataSource.FirstUnsignedPegoutDB()
	if err != nil {
		return SEVERITY_CRITICAL, "", alert.MakeValues(), err
	}

	// No unsigned pegouts
	if unsignedPegout == nil {
		alert.restartsCounter = 0
		return SEVERITY_OK, "OK", alert.MakeValues(), nil
	}

	// Get pegout record from DB
	pegout, err := dataSource.PegoutDB(unsignedPegout.PegoutAddress)
	if err != nil {
		return SEVERITY_CRITICAL, "", alert.MakeValues(), err
	}

	if pegout == nil {
		return SEVERITY_UNKNOWN, "", nil, fmt.Errorf("pegout not found: %s", unsignedPegout.PegoutAddress.String())
	}

	// Check if pegout is new: current is null or new PegoutAddress
	if (alert.currentUnsignedPegout == nil) ||
		(!alert.currentUnsignedPegout.PegoutAddress.Equals(unsignedPegout.PegoutAddress)) {
		// Save new unsigned pegout
		alert.currentUnsignedPegout = unsignedPegout
		alert.restartsCounter = 0

		return SEVERITY_OK, "OK", alert.MakeValues(), nil
	}

	// Check for restart
	if !alert.currentUnsignedPegout.ExpiredAt.Equal(unsignedPegout.ExpiredAt) {
		if !alert.currentUnsignedPegout.ExpiredAt.Equal(time.Unix(0, 0)) {
			alert.restartsCounter++
		}

		alert.currentUnsignedPegout = unsignedPegout
	}

	// Calulate severity
	severity := alert.GetSeverity()
	description := "OK"

	if severity > SEVERITY_OK {
		bitcoinTxId := ""
		if pegout.BitcoinTxId != nil {
			bitcoinTxId = hex.EncodeToString(pegout.BitcoinTxId)
		}

		description = fmt.Sprintf(
			"The pegout signing was restarted %d times.\nPegout: %s.\nBitcoin TX: %s.\nRunbook url: %s",
			alert.restartsCounter,
			mutils.TonExplorerLink(unsignedPegout.PegoutAddress.StringRaw()),
			mutils.BtcExplorerLink(bitcoinTxId),
			mutils.RunbookLink("PegoutRestarts"),
		)
	}

	return severity, Description(description), alert.MakeValues(), nil
}

func (alert *AlertPegoutRestarts) GetSeverity() Severity {
	severity := SEVERITY_OK

	if alert.restartsCounter >= 5 {
		severity = SEVERITY_CRITICAL
	} else if alert.restartsCounter >= 1 {
		severity = SEVERITY_WARNING
	}

	return severity
}

func (alert *AlertPegoutRestarts) MakeValues() Values {
	intValues := make(Values, 1)
	intValues["restarts"] = int64(alert.restartsCounter)

	return intValues
}
