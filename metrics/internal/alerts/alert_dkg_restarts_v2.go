package alerts

import (
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
)

type AlertDkgRestarts struct{}

func NewAlertDkgRestarts() Alert {
	return &AlertDkgRestarts{}
}

func (alert *AlertDkgRestarts) Check(dataSource AlertDataSource) (Severity, Description, Values, error) {
	// Get last DKG Start event
	eventDkgStarted, err := dataSource.Events_Last_DkgStartedDB()
	if err != nil {
		return SEVERITY_CRITICAL, "", alert.makeValues(0), err
	}

	// Chec if no start events
	if eventDkgStarted == nil {
		return SEVERITY_OK, "OK", alert.makeValues(0), nil
	}

	// Get all DKG Restart events after DKG Start event
	eventDkgRestarts, err := dataSource.Events_AllFrom_DkgRestartDB(eventDkgStarted.GetRaw().TxLT)
	if err != nil {
		return SEVERITY_CRITICAL, "", alert.makeValues(0), err
	}

	// Update alert status
	restartsCount := len(eventDkgRestarts)
	severity := alert.getSeverity(restartsCount)
	description := alert.getDescription(severity, restartsCount)
	values := alert.makeValues(restartsCount)

	//for _, restartEvent := range eventDkgRestarts {
	//restartEvent.
	//}

	return severity, description, values, nil
}

func (alert *AlertDkgRestarts) getSeverity(restartsCount int) Severity {
	severity := SEVERITY_OK

	if restartsCount >= 5 {
		severity = SEVERITY_CRITICAL
	} else if restartsCount >= 2 {
		severity = SEVERITY_WARNING
	}

	return severity
}

func (alert *AlertDkgRestarts) getDescription(severity Severity, restartsCount int) Description {
	description := "OK"

	if severity > SEVERITY_OK {
		description = fmt.Sprintf(
			"The DKG was restarted %d times.\n<b>Runbook url:</b> %s",
			restartsCount,
			mutils.RunbookLink("DKGRestarts"),
		)
	}

	return Description(description)
}

func (alert *AlertDkgRestarts) makeValues(restartsCount int) Values {
	values := make(Values, 1)
	values["restarts"] = int64(restartsCount)

	return values
}
