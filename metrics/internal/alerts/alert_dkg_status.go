package alerts

import (
	"fmt"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
)

type AlertDkgStatus struct {
}

func NewAlertDkgStatus() Alert {
	return &AlertDkgStatus{}
}

func (alert *AlertDkgStatus) Check(dataSource AlertDataSource) (Severity, Description, Values, error) {
	// Get DKG
	dkg, err := dataSource.DkgDB()
	if err != nil {
		return SEVERITY_CRITICAL, "", nil, err
	}

	if dkg == nil {
		return Severity(coordinator.DKGStateFinished), "", nil, nil
	}

	description := fmt.Sprintf(
		"The DKG status has changed to %s. Until: %s",
		dkg.State.String(),
		dkg.Until.Format(time.RFC3339),
	)

	return Severity(dkg.State), Description(description), nil, nil
}
