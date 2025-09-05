package alerts

import (
	"context"

	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
)

type AlertDkgEvicted struct{}

func NewAlertDkgEvicted() Alert {
	return &AlertDkgEvicted{}
}

func (alert *AlertDkgEvicted) Check(dataSource AlertDataSource) (Severity, Labels, IntValues, error) {
	// Get DKG
	dkg, err := dataSource.DkgDB()
	if err != nil {
		return SEVERITY_UNKNOWN, nil, nil, err
	}

	if dkg == nil {
		return SEVERITY_OK, nil, nil, nil
	}

	coordinatorContractData, err := dataSource.CoordinatorContractStorageDB()
	if err != nil {
		return SEVERITY_UNKNOWN, nil, nil, err
	}

	vSetSize := len(dkg.VSet)
	validatorsCountMax := 0

	if !coordinatorContractData.StandaloneMode {
		maxValidators, err := dataSource.TonMaxMainValidators(context.Background())
		if err != nil {
			return SEVERITY_UNKNOWN, nil, nil, err
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

	evictedPercentage := mutils.MulDivCeil(uint(evictedCount), 100, uint(count))

	// Calulate severity
	severity := alert.GetSeverity(evictedPercentage)

	return severity, nil, nil, nil
}

func (alert *AlertDkgEvicted) GetSeverity(evictedPercentage uint) Severity {
	severity := SEVERITY_OK

	if evictedPercentage >= 20 {
		severity = SEVERITY_CRITICAL
	} else if evictedPercentage >= 5 {
		severity = SEVERITY_WARNING
	}

	return severity
}
