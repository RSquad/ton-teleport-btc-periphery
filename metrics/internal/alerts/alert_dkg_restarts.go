package alerts

import (
	"time"
)

type AlertDkgRestarts struct {
	dkgUntil        time.Time
	restartsCounter int
}

func NewAlertDkgRestarts() Alert {
	return &AlertDkgRestarts{
		dkgUntil:        time.Unix(0, 0),
		restartsCounter: 0,
	}
}

func (alert *AlertDkgRestarts) Check(dataSource AlertDataSource) (Severity, Labels, error) {
	// Get DKG
	dkg, err := dataSource.DkgDB()
	if err != nil {
		return SEVERITY_UNKNOWN, nil, err
	}

	if dkg == nil {
		alert.dkgUntil = time.Unix(0, 0)
		alert.restartsCounter = 0
		return SEVERITY_OK, nil, nil
	}

	// Check for restart
	if !alert.dkgUntil.Equal(dkg.Until) {
		if alert.dkgUntil != time.Unix(0, 0) {
			alert.restartsCounter++
		}

		alert.dkgUntil = dkg.Until
	}

	// Calulate severity
	severity := alert.GetSeverity(alert.restartsCounter)

	return severity, nil, nil
}

func (alert *AlertDkgRestarts) GetSeverity(restartsCount int) Severity {
	severity := SEVERITY_OK

	if restartsCount >= 10 {
		severity = SEVERITY_CRITICAL
	} else if restartsCount >= 2 {
		severity = SEVERITY_WARNING
	}

	return severity
}
