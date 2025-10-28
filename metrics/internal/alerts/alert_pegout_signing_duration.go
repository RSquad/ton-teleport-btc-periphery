// The alert checks the pegout signing duration. If the signing duration exceeds the limit, the alert fires.
// When the pegout is signed, the alert, if it was fired, keeps firing until there are no unsigned pegouts left
// or the pegout duration falls below the time limit. Severity can increase during the signing process,
// but it can only change after signing. Restarting the signing does not reset the duration timer.

package alerts

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
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

func (alert *AlertPegoutSigningDuration) Check(dataSource AlertDataSource) (Severity, Description, Values, error) {
	// Get first unsigned pegout
	unsignedPegout, err := dataSource.FirstUnsignedPegoutDB()
	if err != nil {
		return SEVERITY_UNKNOWN, "", nil, err
	}

	// No unsigned pegouts
	if unsignedPegout == nil {
		alert.severity = SEVERITY_OK
		alert.currentUnsignedPegout = nil
		return SEVERITY_OK, "OK", nil, nil
	}

	// Get pegout signingTimeout (from Coordinator)
	if alert.signingTimeout == 0 {
		coordinatorData, err := dataSource.CoordinatorContractStorageDB()
		if err != nil {
			return SEVERITY_UNKNOWN, "", nil, err
		}
		alert.signingTimeout = coordinatorData.SigningTimeout
	}

	// Check if pegout is new: current is null or new PegoutAddress
	duration := 0 * time.Second
	if (alert.currentUnsignedPegout == nil) ||
		(!alert.currentUnsignedPegout.PegoutAddress.Equals(unsignedPegout.PegoutAddress)) {
		alert.severity = SEVERITY_OK

		// Save new unsigned pegout
		alert.currentUnsignedPegout = unsignedPegout

		// Calculate beginTimestamp
		expiredAt := unsignedPegout.ExpiredAt.Unix()
		beginTimestamp := dataSource.NowUnixTs()

		if expiredAt > 0 {
			beginTimestamp = expiredAt - int64(alert.signingTimeout)
		}

		if beginTimestamp <= 0 {
			return SEVERITY_UNKNOWN, "", nil, err
		}

		alert.beginTimestamp = beginTimestamp
	} else {
		// Calulate severity
		duration = time.Duration(dataSource.NowUnixTs()-alert.beginTimestamp) * time.Second
		currentSeverity := alert.GetSeverity(duration)

		if currentSeverity > alert.severity {
			alert.severity = currentSeverity
		}
	}

	// Get pegout record from DB
	pegout, err := dataSource.PegoutDB(alert.currentUnsignedPegout.PegoutAddress)
	if err != nil {
		return SEVERITY_UNKNOWN, "", nil, err
	}

	if pegout == nil {
		return SEVERITY_UNKNOWN, "", nil, fmt.Errorf("pegout not found: %s", unsignedPegout.PegoutAddress.String())
	}

	description := "OK"

	if alert.severity > SEVERITY_OK {
		bitcoinTxId := ""
		if pegout.BitcoinTxId != nil {
			bitcoinTxId = hex.EncodeToString(pegout.BitcoinTxId)
		}

		description = fmt.Sprintf(
			"Pegout transaction was not signed within %d minutes. Pegout: %s. Bitcoin TX: %s. Steps to resolve: %s",
			duration/time.Minute,
			mutils.TonExplorerLink(unsignedPegout.PegoutAddress.StringRaw()),
			mutils.BtcExplorerLink(bitcoinTxId),
			mutils.RunbookLink("PegoutSigningDuration"),
		)
	}

	return alert.severity, Description(description), nil, nil
}

func (alert *AlertPegoutSigningDuration) GetSeverity(duration time.Duration) Severity {
	severity := SEVERITY_OK

	if duration >= 22*time.Minute {
		severity = SEVERITY_CRITICAL
	}

	return severity
}
