package alerts

import (
	"encoding/hex"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
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

func (alert *AlertPegoutRestarts) Check(dataSource AlertDataSource) (Severity, Labels, error) {
	labels := Labels{
		"bitcoin_tx_id": "",
		"pegout_addr":   "",
	}

	// Get first unsigned pegout
	unsignedPegout, err := dataSource.FirstUnsignedPegoutDB()
	if err != nil {
		return SEVERITY_UNKNOWN, labels, err
	}

	// No unsigned pegouts
	if unsignedPegout == nil {
		alert.restartsCounter = 0
		return SEVERITY_OK, labels, nil
	}

	// Get pegout record from DB
	pegoutDbRow, err := dataSource.PegoutDB(unsignedPegout.PegoutAddress)
	if err != nil {
		return SEVERITY_UNKNOWN, labels, err
	}

	if pegoutDbRow.BitcoinTxId != nil {
		labels["bitcoin_tx_id"] = hex.EncodeToString(pegoutDbRow.BitcoinTxId)
	}
	labels["pegout_addr"] = unsignedPegout.PegoutAddress.StringRaw()

	// Check if pegout is new: current is null or new PegoutAddress
	if (alert.currentUnsignedPegout == nil) ||
		(!alert.currentUnsignedPegout.PegoutAddress.Equals(unsignedPegout.PegoutAddress)) {
		// Save new unsigned pegout
		alert.currentUnsignedPegout = unsignedPegout
		alert.restartsCounter = 0

		return SEVERITY_OK, labels, nil
	}

	// Check for restart
	if !alert.currentUnsignedPegout.ExpiredAt.Equal(unsignedPegout.ExpiredAt) {
		if alert.currentUnsignedPegout.ExpiredAt != time.Unix(0, 0) {
			alert.restartsCounter++
		}

		alert.currentUnsignedPegout = unsignedPegout
	}

	// Calulate severity
	severity := alert.GetSeverity(alert.restartsCounter)

	return severity, labels, nil
}

func (alert *AlertPegoutRestarts) GetSeverity(restartsCount int) Severity {
	severity := SEVERITY_OK

	if restartsCount >= 10 {
		severity = SEVERITY_CRITICAL
	} else if restartsCount >= 1 {
		severity = SEVERITY_WARNING
	}

	return severity
}
