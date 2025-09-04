package alerts

import (
	"encoding/hex"

	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
)

type AlertPegoutFeeNotReset struct{}

func NewAlertPegoutFeeNotReset() Alert {
	return &AlertPegoutFeeNotReset{}
}

func (alert *AlertPegoutFeeNotReset) Check(dataSource AlertDataSource) (Severity, Labels, error) {
	labels := Labels{
		"bitcoin_tx_id": "",
		"pegout_addr":   "",
	}

	// Get last signed pegout
	pegout, err := dataSource.LastSignedPegoutDB()
	if err != nil {
		return SEVERITY_UNKNOWN, labels, err
	}

	if pegout == nil {
		return SEVERITY_OK, labels, nil
	}

	pegoutBlockHeight := int64(0)
	lastConfirmedBlockHeight := int64(0)
	nextSvb := int64(0)

	if pegout.BitcoinBlockHash == nil {
		return SEVERITY_OK, labels, nil
	}

	// Get info from bitcoin network
	{
		blockHeight, err := dataSource.BtcGetBlockHeightByHash(mutils.BytesToBTCHash(pegout.BitcoinBlockHash))
		if err != nil {
			return SEVERITY_UNKNOWN, labels, err
		}

		pegoutBlockHeight = blockHeight
	}

	// Get last confirmed block
	{
		bitcoinClientContractStorage, err := dataSource.BitcoinClientContractStorageDB()
		if err != nil {
			return SEVERITY_UNKNOWN, labels, err
		}

		lastConfirmedBlockHeight = bitcoinClientContractStorage.LastConfirmedBlockHeight
	}

	// Get next_svb
	{
		storage, err := dataSource.TeleportContractStorageDB()
		if err != nil {
			return SEVERITY_UNKNOWN, labels, err
		}

		nextSvb = int64(storage.NextSVB)
	}

	// Update labels
	if pegout.BitcoinTxId != nil {
		labels["bitcoin_tx_id"] = hex.EncodeToString(pegout.BitcoinTxId)
	}
	labels["pegout_addr"] = pegout.Addr.StringRaw()

	// Calulate severity
	severity := alert.GetSeverity(pegoutBlockHeight, lastConfirmedBlockHeight, nextSvb)

	return severity, labels, nil
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
