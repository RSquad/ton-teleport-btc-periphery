package alerts

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
)

type AlertPegoutCommitments struct{}

func NewAlertPegoutCommitments() Alert {
	return &AlertPegoutCommitments{}
}

func (alert *AlertPegoutCommitments) Check(dataSource AlertDataSource) (Severity, Description, Values, error) {
	component := "AlertPegoutCommitments"
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
		logPegoutCommitmentsFetchError(unsignedPegout.PegoutAddress, err)
		return SEVERITY_UNKNOWN, "", nil, err
	}

	if pegout == nil {
		err := fmt.Errorf("pegout not found: %s", unsignedPegout.PegoutAddress.String())
		logPegoutNotFound(component, unsignedPegout.PegoutAddress, err)
		return SEVERITY_UNKNOWN, "", nil, err
	}

	// Wait until the signing stage starts
	if unsignedPegout.Signatures.Count == 0 {
		logSigningNotStarted(unsignedPegout.PegoutAddress)
		return SEVERITY_OK, "OK", nil, nil
	}

	// Get Prev DKG
	prevDkg, err := dataSource.PrevDkgDB()
	if err != nil {
		logPrevDkgFetchError(component, err)
		return SEVERITY_UNKNOWN, "", nil, err
	}

	// Calculate commitmentsPercentage
	maxSigners := prevDkg.MaxSigners
	commitmentsMask := new(big.Int).Or(
		unsignedPegout.CommitmentsMaskAccepted,
		unsignedPegout.CommitmentsMaskOther,
	)
	commitmentsCount := mutils.Popcnt(commitmentsMask)
	commitmentsPercentage := mutils.MulDivCeil(uint(commitmentsCount), 100, uint(maxSigners))

	// Calculate severity
	severity := alert.GetSeverity(commitmentsPercentage)

	// Update description
	description := "OK"

	bitcoinTxId := ""
	if pegout.BitcoinTxId != nil {
		bitcoinTxId = hex.EncodeToString(pegout.BitcoinTxId)
	}

	if severity > SEVERITY_OK {
		description = fmt.Sprintf(
			"The number of pegout commitments is %d of %d (%d%%).\n<b>Pegout:</b> %s.\n<b>Bitcoin TX:</b> %s.\n<b>Runbook url:</b> %s",
			commitmentsCount,
			maxSigners,
			commitmentsPercentage,
			mutils.TonExplorerLink(unsignedPegout.PegoutAddress.StringRaw()),
			mutils.BtcExplorerLink(bitcoinTxId),
			mutils.RunbookLink("PegoutCommitments"),
		)
	}

	return severity, Description(description), nil, nil
}

func (alert *AlertPegoutCommitments) GetSeverity(commitmentsPercentage uint) Severity {
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
