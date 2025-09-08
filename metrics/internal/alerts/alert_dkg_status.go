package alerts

import (
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
)

type AlertDkgStatus struct {
}

func NewAlertDkgStatus() Alert {
	return &AlertDkgStatus{}
}

func (alert *AlertDkgStatus) Check(dataSource AlertDataSource) (Severity, Labels, IntValues, error) {
	// Get DKG
	dkg, err := dataSource.DkgDB()
	if err != nil {
		return SEVERITY_UNKNOWN, nil, nil, err
	}

	if dkg == nil {
		return Severity(coordinator.DKGStateFinished), nil, nil, nil
	}

	return Severity(dkg.State), nil, nil, nil
}
