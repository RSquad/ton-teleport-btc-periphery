package alerts

import (
	"encoding/hex"

	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
)

type AlertPegoutSigners struct {
}

func NewAlertPegoutSigners() Alert {
	return &AlertPegoutSigners{}
}

func (alert *AlertPegoutSigners) NewLabels() Labels {
	return Labels{
		"bitcoin_tx_id": "",
		"pegout_addr":   "",
	}
}

func (alert *AlertPegoutSigners) Check(dataSource AlertDataSource) (Severity, Labels, Values, error) {
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

	// Calulate signersAllowedPercentage
	maxSigners := unsignedPegout.MaxSigners
	signersAllowedCount := mutils.Popcnt(unsignedPegout.SigningMask)
	signersAllowedPercentage := mutils.MulDivCeil(uint(signersAllowedCount), 100, uint(maxSigners))

	// Calulate severity
	severity := alert.GetSeverity(signersAllowedPercentage)

	return severity, labels, nil, nil
}

func (alert *AlertPegoutSigners) GetSeverity(signersAllowedPercentage uint) Severity {
	severity := SEVERITY_OK

	if signersAllowedPercentage <= 70 {
		severity = SEVERITY_CRITICAL
	} else if signersAllowedPercentage <= 80 {
		severity = SEVERITY_WARNING
	} else if signersAllowedPercentage <= 90 {
		severity = SEVERITY_INFO
	}

	return severity
}
