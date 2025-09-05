package alerts

import (
	"context"

	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
)

type AlertDkgParticipants struct{}

func NewAlertDkgParticipants() Alert {
	return &AlertDkgParticipants{}
}

func (alert *AlertDkgParticipants) Check(dataSource AlertDataSource) (Severity, Labels, IntValues, error) {
	// Get PrevDKG
	prevDkg, err := dataSource.PrevDkgDB()
	if err != nil {
		return SEVERITY_UNKNOWN, nil, nil, err
	}

	if prevDkg == nil {
		return SEVERITY_OK, nil, nil, nil
	}

	// Calulate participantsPercentage
	maxParticipants, err := dataSource.TonMaxMainValidators(context.Background())
	if err != nil {
		return SEVERITY_UNKNOWN, nil, nil, err
	}

	participantsCount := prevDkg.MaxSigners
	participantsPercentage := mutils.MulDivCeil(uint(participantsCount), 100, uint(maxParticipants))

	// Calulate severity
	severity := alert.GetSeverity(participantsPercentage)

	return severity, nil, nil, nil
}

func (alert *AlertDkgParticipants) GetSeverity(participantsPercentage uint) Severity {
	severity := SEVERITY_OK

	if participantsPercentage <= 55 {
		severity = SEVERITY_CRITICAL
	} else if participantsPercentage <= 80 {
		severity = SEVERITY_WARNING
	}

	return severity
}
