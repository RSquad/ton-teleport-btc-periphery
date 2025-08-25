package alerts

import (
	"errors"
	"math/big"

	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
)

type AlertPegoutMaxSignersCount struct {
}

func NewAlertPegoutMaxSignersCount() Alert {
	return &AlertPegoutMaxSignersCount{}
}

func (alert *AlertPegoutMaxSignersCount) Check(dataSource AlertDataSource) (Severity, []string, error) {
	// Get first unsigned pegout
	unsignedPegout, err := dataSource.FirstUnsignedPegout()
	if err != nil {
		return SEVERITY_UNKNOWN, nil, err
	}

	// No unsigned pegouts
	if unsignedPegout == nil {
		return SEVERITY_OK, nil, nil
	}

	// Get prev DKG
	prevDkg, err := dataSource.PrevDkg()
	if err != nil {
		return SEVERITY_UNKNOWN, nil, err
	}

	if prevDkg == nil {
		return SEVERITY_UNKNOWN, []string{}, errors.New("PrevDKG is null")
	}

	// Calulate commitmentsPercentage
	maxSigners := prevDkg.MaxSigners
	commitmentsMask := new(big.Int).Or(
		unsignedPegout.CommitmentsMaskAccepted,
		unsignedPegout.CommitmentsMaskOther,
	)
	commitmentsCount := mutils.Popcnt(commitmentsMask)
	commitmentsPercentage := mutils.MulDivCeil(uint(commitmentsCount), 100, uint(maxSigners))

	// Calulate severity
	severity := alert.GetSeverity(commitmentsPercentage)

	return severity, []string{}, nil
}

func (alert *AlertPegoutMaxSignersCount) GetSeverity(commitmentsPercentage uint) Severity {
	severity := SEVERITY_OK

	if commitmentsPercentage <= 70 {
		severity = SEVERITY_CRITICAL
	} else if commitmentsPercentage <= 80 {
		severity = SEVERITY_WARNING
	} else if commitmentsPercentage <= 90 {
		severity = SEVERITY_INFO
	}

	return severity
}
