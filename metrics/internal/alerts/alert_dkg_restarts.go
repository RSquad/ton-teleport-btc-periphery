package alerts

import (
	"fmt"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
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

func (alert *AlertDkgRestarts) Check(dataSource AlertDataSource) (Severity, Description, Values, error) {
	// Get DKG
	dkg, err := dataSource.DkgDB()
	if err != nil {
		return SEVERITY_CRITICAL, "", alert.MakeValues(), err
	}

	if dkg == nil {
		alert.dkgUntil = time.Unix(0, 0)
		alert.restartsCounter = 0
		return SEVERITY_OK, "OK", alert.MakeValues(), nil
	}

	if dkg.State == coordinator.DKGStateFinished {
		alert.dkgUntil = time.Unix(0, 0)
		alert.restartsCounter = 0
		return SEVERITY_OK, "OK", alert.MakeValues(), nil
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
	description := "OK"

	if severity > SEVERITY_OK {
		description = fmt.Sprintf(
			"The DKG was restarted %d times. Steps to resolve: %s",
			alert.restartsCounter,
			mutils.RunbookLink("DKGRestarts"),
		)
	}

	return severity, Description(description), alert.MakeValues(), nil
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

func (alert *AlertDkgRestarts) MakeValues() Values {
	values := make(Values, 1)
	values["restarts"] = int64(alert.restartsCounter)

	return values
}
