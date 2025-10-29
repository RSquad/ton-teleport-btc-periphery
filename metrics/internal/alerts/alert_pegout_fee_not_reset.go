package alerts

import (
	"encoding/hex"
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
	"github.com/xssnick/tonutils-go/address"
)

type AlertPegoutFeeNotReset struct{}

func NewAlertPegoutFeeNotReset() Alert {
	return &AlertPegoutFeeNotReset{}
}

func (alert *AlertPegoutFeeNotReset) Check(dataSource AlertDataSource) (Severity, Description, Values, error) {
	// Get last signed pegout
	pegout, err := dataSource.LastSignedPegoutDB()
	if err != nil {
		return SEVERITY_CRITICAL, "", nil, err
	}

	if pegout == nil {

		// TODO: add LastConfirmedPegout...

		return SEVERITY_OK, "OK", nil, nil
	}

	pegoutBlockHeight := int64(0)
	lastConfirmedBlockHeight := int64(0)
	nextSvb := int64(0)

	if pegout.BitcoinBlockHash == nil {
		return SEVERITY_OK, "OK", nil, nil
	}

	// Get info from bitcoin network
	{
		blockHeight, err := dataSource.BtcGetBlockHeightByHash(mutils.BytesToBTCHash(pegout.BitcoinBlockHash))
		if err != nil {
			return SEVERITY_CRITICAL, "", nil, err
		}

		pegoutBlockHeight = blockHeight
	}

	// Get last confirmed block
	{
		bitcoinClientContractStorage, err := dataSource.BitcoinClientContractStorageDB()
		if err != nil {
			return SEVERITY_CRITICAL, "", nil, err
		}

		lastConfirmedBlockHeight = bitcoinClientContractStorage.LastConfirmedBlockHeight
	}

	// Get next_svb
	{
		storage, err := dataSource.TeleportContractStorageDB()
		if err != nil {
			return SEVERITY_CRITICAL, "", nil, err
		}

		nextSvb = int64(storage.NextSVB)
	}

	// Calulate severity
	severity := alert.GetSeverity(pegoutBlockHeight, lastConfirmedBlockHeight, nextSvb)
	description := "OK"

	if severity > SEVERITY_OK {
		bitcoinTxId := ""
		if pegout.BitcoinTxId != nil {
			bitcoinTxId = hex.EncodeToString(pegout.BitcoinTxId)
		}

		description = fmt.Sprintf(
			"Pegout transaction already has %d (pegoutBlockHeight %d, lastConfirmedBlockHeight %d) confirmations but the fee has not been reset (nextSvb = %d).\nPegout: %s.\nBitcoin TX: %s.\nRunbook url: %s",
			lastConfirmedBlockHeight-pegoutBlockHeight,
			pegoutBlockHeight,
			lastConfirmedBlockHeight,
			nextSvb,
			mutils.TonExplorerLink((*address.Address)(pegout.Addr).StringRaw()),
			mutils.BtcExplorerLink(bitcoinTxId),
			mutils.RunbookLink("PegoutFeeNotReset"),
		)
	}

	return severity, Description(description), nil, nil
}

func (alert *AlertPegoutFeeNotReset) GetSeverity(
	pegoutBlockHeight int64,
	lastConfirmedBlockHeight int64,
	nextSvb int64,
) Severity {
	severity := SEVERITY_OK

	if (pegoutBlockHeight < lastConfirmedBlockHeight) && (nextSvb > 0) {
		severity = SEVERITY_CRITICAL
	}

	return severity
}
