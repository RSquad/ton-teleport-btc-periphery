package alerts

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
)

type AlertPegoutCommintments struct {
}

func NewAlertPegoutCommintments() Alert {
	return &AlertPegoutCommintments{}
}

func (alert *AlertPegoutCommintments) Check(dataSource AlertDataSource) (Severity, Description, Values, error) {
	// Get first unsigned pegout
	unsignedPegout, err := dataSource.FirstUnsignedPegoutDB()
	if err != nil {
		return SEVERITY_UNKNOWN, "", nil, err
	}

	// No unsigned pegouts
	if unsignedPegout == nil {
		return SEVERITY_OK, "OK", nil, nil
	}

	// Get pegout record from DB
	pegout, err := dataSource.PegoutDB(unsignedPegout.PegoutAddress)
	if err != nil {
		return SEVERITY_UNKNOWN, "", nil, err
	}

	// Wait until the signing stage starts
	if unsignedPegout.Signatures.Count == 0 {
		return SEVERITY_OK, "OK", nil, nil
	}

	// Get Prev DKG
	prevDkg, err := dataSource.PrevDkgDB()
	if err != nil {
		return SEVERITY_UNKNOWN, "", nil, err
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

	// Update description
	description := "OK"

	if severity > SEVERITY_OK {
		bitcoinTxId := ""
		if pegout.BitcoinTxId != nil {
			bitcoinTxId = hex.EncodeToString(pegout.BitcoinTxId)
		}

		description = fmt.Sprintf(
			"The number of pegout commitments is %d of %d (%d%%). Pegout: %s. Bitcoin TX: %s",
			commitmentsCount,
			maxSigners,
			commitmentsPercentage,
			mutils.TonExplorerLink(unsignedPegout.PegoutAddress.StringRaw()),
			mutils.BtcExplorerLink(bitcoinTxId),
		)
	}

	return severity, Description(description), nil, nil
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
