// The alert checks the pegout signing duration. If the signing duration exceeds the limit, the alert fires.
// When the pegout is signed, the alert, if it was fired, keeps firing until there are no unsigned pegouts left
// or the pegout duration falls below the time limit. Severity can increase during the signing process,
// but it can only change after signing. Restarting the signing does not reset the duration timer.

package alerts

import (
	"encoding/hex"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
)

type AlertPegoutSigningDuration struct {
	currentUnsignedPegout *coordinator.PegoutRecord
	signingTimeout        uint32
	beginTimestamp        int64
	severity              Severity
}

func NewAlertPegoutSigningDuration() Alert {
	return &AlertPegoutSigningDuration{
		currentUnsignedPegout: nil,
		signingTimeout:        0,
		beginTimestamp:        0,
		severity:              SEVERITY_OK,
	}
}

func (alert *AlertPegoutSigningDuration) Check(dataSource AlertDataSource) (Severity, Labels, error) {
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
		alert.severity = SEVERITY_OK
		alert.currentUnsignedPegout = nil
		return SEVERITY_OK, labels, nil
	}

	// Get pegout signingTimeout (from Coordinator)
	if alert.signingTimeout == 0 {
		coordinatorData, err := dataSource.CoordinatorContractStorageDB()
		if err != nil {
			return SEVERITY_UNKNOWN, labels, err
		}
		alert.signingTimeout = coordinatorData.SigningTimeout
	}

	// Check if pegout is new: current is null or new PegoutAddress
	if (alert.currentUnsignedPegout == nil) ||
		(!alert.currentUnsignedPegout.PegoutAddress.Equals(unsignedPegout.PegoutAddress)) {
		// Update severity
		if alert.currentUnsignedPegout != nil {
			// alert.beginTimestamp - from previous pegout
			duration := time.Duration(dataSource.NowUnixTs()-alert.beginTimestamp) * time.Second

			// alert.severity will be calculated from previous pegout
			alert.severity = alert.GetSeverity(duration)
		}

		// Save new unsigned pegout
		alert.currentUnsignedPegout = unsignedPegout

		// Calculate beginTimestamp
		expiredAt := unsignedPegout.ExpiredAt.Unix()
		beginTimestamp := dataSource.NowUnixTs()

		if expiredAt > 0 {
			beginTimestamp = expiredAt - int64(alert.signingTimeout)
		}

		if beginTimestamp <= 0 {
			return SEVERITY_UNKNOWN, labels, err
		}

		alert.beginTimestamp = beginTimestamp
	} else {

		// Calulate severity
		duration := time.Duration(dataSource.NowUnixTs()-alert.beginTimestamp) * time.Second
		currentSeverity := alert.GetSeverity(duration)

		if currentSeverity > alert.severity {
			alert.severity = currentSeverity
		}
	}

	// Get pegout record from DB
	pegout, err := dataSource.PegoutDB(alert.currentUnsignedPegout.PegoutAddress)
	if err != nil {
		return SEVERITY_UNKNOWN, labels, err
	}

	// Update labels
	if pegout.BitcoinTxId != nil {
		labels["bitcoin_tx_id"] = hex.EncodeToString(pegout.BitcoinTxId)
	}
	labels["pegout_addr"] = alert.currentUnsignedPegout.PegoutAddress.StringRaw()

	return alert.severity, labels, nil
}

func (alert *AlertPegoutSigningDuration) GetSeverity(duration time.Duration) Severity {
	severity := SEVERITY_OK

	if duration >= 20*time.Minute {
		severity = SEVERITY_CRITICAL
	}

	return severity
}
