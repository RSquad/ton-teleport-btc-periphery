package alerts

import (
	"encoding/hex"
	"math/big"

	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
)

type AlertPegoutCommintments struct {
}

func NewAlertPegoutCommintments() Alert {
	return &AlertPegoutCommintments{}
}

func (alert *AlertPegoutCommintments) NewLabels() Labels {
	return Labels{
		"bitcoin_tx_id": "",
		"pegout_addr":   "",
		"threshold":     "", // TODO: add threshold value
	}
}

func (alert *AlertPegoutCommintments) Check(dataSource AlertDataSource) (Severity, Labels, Values, error) {
	labels := alert.NewLabels()

	// Get first unsigned pegout
	unsignedPegout, err := dataSource.FirstUnsignedPegoutDB()
	if err != nil {
		return SEVERITY_UNKNOWN, labels, nil, err
	}

	// No unsigned pegouts
	if unsignedPegout == nil {
		return SEVERITY_OK, labels, nil, nil
	}

	// Get pegout record from DB
	pegout, err := dataSource.PegoutDB(unsignedPegout.PegoutAddress)
	if err != nil {
		return SEVERITY_UNKNOWN, labels, nil, err
	}

	// Update labels
	if pegout.BitcoinTxId != nil {
		labels["bitcoin_tx_id"] = hex.EncodeToString(pegout.BitcoinTxId)
	}
	labels["pegout_addr"] = unsignedPegout.PegoutAddress.StringRaw()

	// Wait until the signing stage starts
	if unsignedPegout.Signatures.Count == 0 {
		return SEVERITY_OK, labels, nil, nil
	}

	// Calulate commitmentsPercentage
	maxSigners := unsignedPegout.MaxSigners
	commitmentsMask := new(big.Int).Or(
		unsignedPegout.CommitmentsMaskAccepted,
		unsignedPegout.CommitmentsMaskOther,
	)
	commitmentsCount := mutils.Popcnt(commitmentsMask)
	commitmentsPercentage := mutils.MulDivCeil(uint(commitmentsCount), 100, uint(maxSigners))

	// Calulate severity
	severity := alert.GetSeverity(commitmentsPercentage)

	return severity, labels, nil, nil
}

func (alert *AlertPegoutCommintments) GetSeverity(commitmentsPercentage uint) Severity {
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
