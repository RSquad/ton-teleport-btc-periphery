package alerts

import (
	"context"
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
)

type AlertDkgParticipants struct {
	severity    Severity
	description Description
}

func NewAlertDkgParticipants() Alert {
	return &AlertDkgParticipants{
		severity:    SEVERITY_UNKNOWN,
		description: "",
	}
}

func (alert *AlertDkgParticipants) Check(dataSource AlertDataSource) (Severity, Description, Values, error) {
	// Get DKG
	dkg, err := dataSource.DkgDB()
	if err != nil {
		return SEVERITY_CRITICAL, "", nil, err
	}

	if dkg == nil {
		return alert.severity, "OK", nil, nil
	}

	if dkg.State != coordinator.DKGStateFinished {
		return alert.severity, alert.description, nil, nil
	}

	coordinatorContractData, err := dataSource.CoordinatorContractStorageDB()
	if err != nil {
		return SEVERITY_CRITICAL, "", nil, err
	}

	vSetSize := len(dkg.VSet)
	validatorsCountMax := 0

	if !coordinatorContractData.StandaloneMode {
		maxValidators, err := dataSource.TonMaxMainValidators(context.Background())
		if err != nil {
			return SEVERITY_CRITICAL, "", nil, err
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
	alert.description = Description(fmt.Sprintf(
		"The number of DKG participants is %d of %d (%d%%)",
		count-evictedCount,
		count,
		percentage,
	))
	if alert.severity > SEVERITY_OK {
		alert.description = Description(fmt.Sprintf(
			"The number of DKG participants is %d of %d (%d%%). Runbook url: %s",
			count-evictedCount,
			count,
			percentage,
			mutils.RunbookLink("DKGParticipants"),
		))
	}

	return alert.severity, alert.description, nil, nil
}

func (alert *AlertDkgParticipants) GetSeverity(percentage uint) Severity {
	severity := SEVERITY_OK

	if percentage <= 65 {
		severity = SEVERITY_CRITICAL
	} else if percentage <= 80 {
		severity = SEVERITY_WARNING
	}

	return severity
}
