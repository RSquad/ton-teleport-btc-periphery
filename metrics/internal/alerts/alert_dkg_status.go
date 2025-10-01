package alerts

import (
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
)

type AlertDkgStatus struct {
}

func NewAlertDkgStatus() Alert {
	return &AlertDkgStatus{}
}

func (alert *AlertDkgStatus) NewLabels() Labels {
	return Labels{
		"until": "",
	}
}

func (alert *AlertDkgStatus) Check(dataSource AlertDataSource) (Severity, Labels, Values, error) {
	labels := alert.NewLabels()

	// Get DKG
	dkg, err := dataSource.DkgDB()
	if err != nil {
		return SEVERITY_UNKNOWN, labels, nil, err
	}

	if dkg == nil {
		return Severity(coordinator.DKGStateFinished), labels, nil, nil
	}

	labels["until"] = dkg.Until.Format(time.RFC3339)

	return Severity(dkg.State), labels, nil, nil
}
