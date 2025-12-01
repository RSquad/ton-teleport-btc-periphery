package alerts

import (
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
)

type AlertDkgRestarts struct{}

type info struct {
	restartsCount int
	culpritsCount int
}

func newInfo() *info {
	return &info{
		restartsCount: 0,
		culpritsCount: 0,
	}
}

func NewAlertDkgRestarts() Alert {
	return &AlertDkgRestarts{}
}

func (alert *AlertDkgRestarts) Check(dataSource AlertDataSource) (Severity, Description, Values, error) {
	// Get last DKG Start event
	eventDkgStarted, err := dataSource.Events_Last_DkgStartedDB()
	if err != nil {
		return SEVERITY_CRITICAL, "", alert.makeValues(newInfo()), err
	}

	// Chec if no start events
	if eventDkgStarted == nil {
		return SEVERITY_OK, "OK", alert.makeValues(newInfo()), nil
	}

	// Get all DKG Restart events after DKG Start event
	eventDkgRestarts, err := dataSource.Events_AllFrom_DkgRestartDB(eventDkgStarted.GetRaw().TxLT)
	if err != nil {
		return SEVERITY_CRITICAL, "", alert.makeValues(newInfo()), err
	}

	// Update alert status
	info := alert.getInfo(eventDkgRestarts, eventDkgStarted)
	severity := alert.getSeverity(&info)
	description := alert.getDescription(severity, &info)
	values := alert.makeValues(&info)

	return severity, description, values, nil
}

func (alert *AlertDkgRestarts) getInfo(eventDkgRestarts []*coordinator.DKGRestartedEvent, eventDkgStart *coordinator.DKGStartedEvent) info {
	restartsCount := len(eventDkgRestarts)

	// e := eventDkgRestarts[0]
	// TODO: implement

	return info{
		restartsCount: restartsCount,
		culpritsCount: 0,
	}
}

func (alert *AlertDkgRestarts) getSeverity(inf *info) Severity {
	severity := SEVERITY_OK

	if (inf.restartsCount >= 5) || (inf.culpritsCount > 0) {
		severity = SEVERITY_CRITICAL
	} else if inf.restartsCount >= 2 {
		severity = SEVERITY_WARNING
	}

	return severity
}

func (alert *AlertDkgRestarts) getDescription(severity Severity, inf *info) Description {
	description := "OK"

	if severity > SEVERITY_OK {
		description = fmt.Sprintf(
			"The DKG was restarted %d times.\n<b>Runbook url:</b> %s",
			inf.restartsCount,
			mutils.RunbookLink("DKGRestarts"),
		)
	}

	return Description(description)
}

func (alert *AlertDkgRestarts) makeValues(inf *info) Values {
	values := make(Values, 3)
	values["restarts"] = int64(inf.restartsCount)
	// values["evicted_ids"] = culpritIds
	// values["culprit_ids"] = culpritIds

	return values
}
