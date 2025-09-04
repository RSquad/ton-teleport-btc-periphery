package alerts

import (
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
)

type AlertDkgStatus struct {
}

func NewAlertDkgStatus() Alert {
	return &AlertPegoutRestarts{}
}

func (alert *AlertDkgStatus) Check(dataSource AlertDataSource) (Severity, Labels, error) {
	// Get DKG
	dkg, err := dataSource.DkgDB()
	if err != nil {
		return SEVERITY_UNKNOWN, nil, err
	}

	if dkg == nil {
		return Severity(coordinator.DKGStateFinished), nil, nil
	}

	return Severity(dkg.State), nil, nil
}
