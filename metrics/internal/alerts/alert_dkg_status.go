package alerts

import (
	"fmt"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
)

type AlertDkgStatus struct{}

func NewAlertDkgStatus() Alert {
	return &AlertDkgStatus{}
}

func (alert *AlertDkgStatus) Check(dataSource AlertDataSource) (Severity, Description, Values, error) {
	// Get DKG
	dkg, err := dataSource.DkgDB()
	if err != nil {
		logDkgFetchError(err)
		return SEVERITY_CRITICAL, "", nil, err
	}

	if dkg == nil {
		logNoDkgFound()
		return Severity(coordinator.DKGStateFinished), "", nil, nil
	}

	description := fmt.Sprintf(
		"The DKG status has changed to %s. Until: %s\n<b>Runbook url:</b> %s",
		dkg.State.String(),
		dkg.Until.Format(time.RFC3339),
		mutils.RunbookLink("DKGStatus"),
	)

	severity := Severity(dkg.State)

	return severity, Description(description), nil, nil
}
