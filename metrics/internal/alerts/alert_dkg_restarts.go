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

func (alert *AlertDkgRestarts) Check(dataSource AlertDataSource) (Severity, Labels, IntValues, error) {
	// Get DKG
	dkg, err := dataSource.DkgDB()
	if err != nil {
		return SEVERITY_UNKNOWN, nil, alert.MakeIntValues(), err
	}

	if dkg == nil {
		alert.dkgUntil = time.Unix(0, 0)
		alert.restartsCounter = 0
		return SEVERITY_OK, nil, alert.MakeIntValues(), nil
	}

	// Check for restart
	if !alert.dkgUntil.Equal(dkg.Until) {
		if alert.dkgUntil != time.Unix(0, 0) {
			alert.restartsCounter++
		}

		alert.dkgUntil = dkg.Until
	}

	// Calulate severity
	severity := alert.GetSeverity()

	return severity, nil, alert.MakeIntValues(), nil
}

func (alert *AlertDkgRestarts) GetSeverity() Severity {
	severity := SEVERITY_OK

	if alert.restartsCounter >= 5 {
		severity = SEVERITY_CRITICAL
	} else if alert.restartsCounter >= 2 {
		severity = SEVERITY_WARNING
	}

	return severity
}

func (alert *AlertDkgRestarts) MakeIntValues() IntValues {
	intValues := make(IntValues, 1)
	intValues["restarts"] = int64(alert.restartsCounter)

	return intValues
}
