package alerts

import (
	"encoding/hex"
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
)

type AlertPegoutSigners struct{}

func NewAlertPegoutSigners() Alert {
	return &AlertPegoutSigners{}
}

func (alert *AlertPegoutSigners) Check(dataSource AlertDataSource) (Severity, Description, Values, error) {
	component := "AlertPegoutSigners"

	// Get first unsigned pegout
	unsignedPegout, err := dataSource.FirstUnsignedPegoutDB()
	if err != nil {
		logUnsignedPegoutFetchError(component, err)
		return SEVERITY_UNKNOWN, "", nil, err
	}

	// No unsigned pegouts
	if unsignedPegout == nil {
		logNoUnsignedPegouts(component)
		return SEVERITY_OK, "OK", nil, nil
	}

	// Get pegout record from DB
	pegout, err := dataSource.PegoutDB(unsignedPegout.PegoutAddress)
	if err != nil {
		logPegoutFetchError(component, err)
		return SEVERITY_UNKNOWN, "", nil, err
	}

	if pegout == nil {
		err := fmt.Errorf("pegout not found: %s", unsignedPegout.PegoutAddress.String())
		logPegoutNotFound(component, unsignedPegout.PegoutAddress, err)
		return SEVERITY_UNKNOWN, "", nil, err
	}

	// Get Prev DKG
	prevDkg, err := dataSource.PrevDkgDB()
	if err != nil {
		logPrevDkgFetchError(component, err)
		return SEVERITY_UNKNOWN, "", nil, err
	}

	// Calculate signersAllowedPercentage
	maxSigners := prevDkg.MaxSigners
	signersAllowedCount := mutils.Popcnt(unsignedPegout.SigningMask)
	signersAllowedPercentage := mutils.MulDivCeil(uint(signersAllowedCount), 100, uint(maxSigners))

	// Calculate severity
	severity := alert.GetSeverity(signersAllowedPercentage)
	description := "OK"

	bitcoinTxId := ""
	if pegout.BitcoinTxId != nil {
		bitcoinTxId = hex.EncodeToString(pegout.BitcoinTxId)
	}

	if severity > SEVERITY_OK {
		description = fmt.Sprintf(
			"Number of validators allowed to sign pegout is %d of %d (%d%%).\n<b>Pegout:</b> %s.\n<b>Bitcoin TX:</b> %s.\n<b>Runbook url:</b> %s",
			signersAllowedCount,
			maxSigners,
			signersAllowedPercentage,
			mutils.TonExplorerLink(unsignedPegout.PegoutAddress.StringRaw()),
			mutils.BtcExplorerLink(bitcoinTxId),
			mutils.RunbookLink("PegoutSigners"),
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
