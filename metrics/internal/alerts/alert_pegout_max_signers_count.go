package alerts

import (
	"errors"
	"math/big"

	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/utils"
)

type AlertPegoutMaxSignersCount struct {
}

func NewAlertPegoutMaxSignersCount() Alert {
	return &AlertPegoutMaxSignersCount{}
}

func (alert *AlertPegoutMaxSignersCount) Check(dataSource AlertDataSource) (Severity, error) {
	// Get first unsigned pegout
	unsignedPegout, err := dataSource.FirstUnsignedPegout()
	if err != nil {
		return SEVERITY_OK, err
	}

	// No unsigned pegouts
	if unsignedPegout == nil {
		return SEVERITY_OK, nil
	}

	// Get prev DKG
	prevDkg, err := dataSource.PrevDkg()
	if err != nil {
		return SEVERITY_OK, err
	}

	if prevDkg == nil {
		return SEVERITY_OK, errors.New("PrevDKG is null")
	}

	// Calulate commitmentsPercentage
	maxSigners := prevDkg.MaxSigners
	commitmentsMask := new(big.Int).Or(
		unsignedPegout.CommitmentsMaskAccepted,
		unsignedPegout.CommitmentsMaskOther,
	)
	commitmentsCount := utils.Popcnt(commitmentsMask)
	commitmentsPercentage := utils.MulDivCeil(uint(commitmentsCount), 100, uint(maxSigners))

	// Calulate severity
	severity := SEVERITY_OK

	if commitmentsPercentage <= 70 {
		severity = SEVERITY_CRITICAL
	} else if commitmentsPercentage <= 80 {
		severity = SEVERITY_WARNING
	} else if commitmentsPercentage <= 90 {
		severity = SEVERITY_INFO
	}

	return severity, nil
}
