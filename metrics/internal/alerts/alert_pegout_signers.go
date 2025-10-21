package alerts

import (
	"encoding/hex"
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
)

type AlertPegoutSigners struct {
}

func NewAlertPegoutSigners() Alert {
	return &AlertPegoutSigners{}
}

func (alert *AlertPegoutSigners) Check(dataSource AlertDataSource) (Severity, Description, Values, error) {
	// Get first unsigned pegout
	unsignedPegout, err := dataSource.FirstUnsignedPegoutDB()
	if err != nil {
		return SEVERITY_CRITICAL, "", nil, err
	}

	// No unsigned pegouts
	if unsignedPegout == nil {
		return SEVERITY_OK, "OK", nil, nil
	}

	// Get pegout record from DB
	pegout, err := dataSource.PegoutDB(unsignedPegout.PegoutAddress)
	if err != nil {
		return SEVERITY_CRITICAL, "", nil, err
	}

	// Calulate signersAllowedPercentage
	maxSigners := unsignedPegout.MaxSigners
	signersAllowedCount := mutils.Popcnt(unsignedPegout.SigningMask)
	signersAllowedPercentage := mutils.MulDivCeil(uint(signersAllowedCount), 100, uint(maxSigners))

	// Calulate severity
	severity := alert.GetSeverity(signersAllowedPercentage)
	description := "OK"

	if severity > SEVERITY_OK {
		bitcoinTxId := ""
		if pegout.BitcoinTxId != nil {
			bitcoinTxId = hex.EncodeToString(pegout.BitcoinTxId)
		}

		description = fmt.Sprintf(
			"Number of validators allowed to sign pegout is %d of %d (%d%%). Pegout: %s. Bitcoin TX: %s",
			signersAllowedCount,
			maxSigners,
			signersAllowedPercentage,
			mutils.TonExplorerLink(unsignedPegout.PegoutAddress.StringRaw()),
			mutils.BtcExplorerLink(bitcoinTxId),
		)
	}

	return severity, Description(description), nil, nil
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
