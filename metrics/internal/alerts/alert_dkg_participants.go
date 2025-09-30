package alerts

import (
	"context"
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
)

type AlertDkgParticipants struct {
	severity Severity
	labels   Labels
}

func NewAlertDkgParticipants() Alert {
	return &AlertDkgParticipants{
		severity: SEVERITY_UNKNOWN,
		labels:   Labels{},
	}
}

func (alert *AlertDkgParticipants) NewLabels() Labels {
	return Labels{
		"participants_count": "",
	}
}

func (alert *AlertDkgParticipants) Check(dataSource AlertDataSource) (Severity, Labels, IntValues, error) {
	// Get DKG
	dkg, err := dataSource.DkgDB()
	if err != nil {
		return SEVERITY_UNKNOWN, alert.NewLabels(), nil, err
	}

	if dkg == nil {
		return alert.severity, alert.labels, nil, nil
	}

	if dkg.State != coordinator.DKGStateFinished {
		return alert.severity, alert.labels, nil, nil
	}

	coordinatorContractData, err := dataSource.CoordinatorContractStorageDB()
	if err != nil {
		return SEVERITY_UNKNOWN, alert.NewLabels(), nil, err
	}

	vSetSize := len(dkg.VSet)
	validatorsCountMax := 0

	if !coordinatorContractData.StandaloneMode {
		maxValidators, err := dataSource.TonMaxMainValidators(context.Background())
		if err != nil {
			return SEVERITY_UNKNOWN, alert.NewLabels(), nil, err
		}

		validatorsCountMax = maxValidators
	} else {
		validatorsCountMax = vSetSize
	}

	count := min(vSetSize, validatorsCountMax)
	evictedCount := 0

	for i := range count {
		if dkg.VSetMask.Bit(i) == 0 {
			evictedCount++
		}
	}

	percentage := mutils.MulDivCeil(uint(count-evictedCount), 100, uint(count))

	// Calulate severity
	alert.severity = alert.GetSeverity(percentage)
	alert.labels["participants_count"] = fmt.Sprintf("%d of %d (%d%%)", count-evictedCount, count, percentage)

	return alert.severity, alert.labels, nil, nil
}

func (alert *AlertDkgParticipants) GetSeverity(percentage uint) Severity {
	severity := SEVERITY_OK

	if percentage <= 55 {
		severity = SEVERITY_CRITICAL
	} else if percentage <= 80 {
		severity = SEVERITY_WARNING
	}

	return severity
}
