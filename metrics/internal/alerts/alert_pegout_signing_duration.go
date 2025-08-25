// The alert checks the pegout signing duration. If the signing duration exceeds the limit, the alert fires.
// When the pegout is signed, the alert, if it was fired, keeps firing until there are no unsigned pegouts left
// or the pegout duration falls below the time limit. Severity can increase during the signing process,
// but it can only change after signing. Restarting the signing does not reset the duration timer.

package alerts

import (
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

func (alert *AlertPegoutSigningDuration) Check(dataSource AlertDataSource) (Severity, []string, error) {
	// Get first unsigned pegout
	unsignedPegout, err := dataSource.FirstUnsignedPegout()
	if err != nil {
		return SEVERITY_UNKNOWN, nil, err
	}

	// No unsigned pegouts
	if unsignedPegout == nil {
		alert.severity = SEVERITY_OK
		return SEVERITY_OK, nil, nil
	}

	// Get pegout signingTimeout (from Coordinator)
	if alert.signingTimeout == 0 {
		coordinatorData, err := dataSource.CoordinatorContractData()
		if err != nil {
			return SEVERITY_UNKNOWN, nil, err
		}
		alert.signingTimeout = coordinatorData.SigningTimeout
	}

	// Check if pegout is new: current is null or new PegoutAddress
	if (alert.currentUnsignedPegout == nil) ||
		(!alert.currentUnsignedPegout.PegoutAddress.Equals(unsignedPegout.PegoutAddress)) {
		// Update severity
		if alert.currentUnsignedPegout != nil {
			duration := time.Duration(dataSource.NowUnixTs()-alert.beginTimestamp) * time.Second
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
			return SEVERITY_UNKNOWN, nil, err
		}

		alert.beginTimestamp = beginTimestamp
		return alert.severity, nil, nil
	}

	// Calulate severity
	duration := time.Duration(dataSource.NowUnixTs()-alert.beginTimestamp) * time.Second
	currentSeverity := alert.GetSeverity(duration)

	if currentSeverity > alert.severity {
		alert.severity = currentSeverity
	}

	return alert.severity, []string{}, nil
}

func (alert *AlertPegoutSigningDuration) GetSeverity(duration time.Duration) Severity {
	severity := SEVERITY_OK

	if duration >= 20*time.Minute {
		severity = SEVERITY_CRITICAL
	} else if duration >= 10*time.Minute {
		severity = SEVERITY_WARNING
	}

	return severity
}
