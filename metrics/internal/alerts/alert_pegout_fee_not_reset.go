package alerts

import (
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/bitcoinclientcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
)

type AlertPegoutFeeNotReset struct {
	// TODO: move to AlertDataSource
	bitcoinClient         *bitcoin.Client
	bitcoinClientContract *bitcoinclientcontract.BitcoinClientContract
	teleportContract      *teleportcontract.TeleportContract
}

func NewAlertPegoutFeeNotReset(
	// TODO: move to AlertDataSource
	bitcoinClient *bitcoin.Client,
	bitcoinClientContract *bitcoinclientcontract.BitcoinClientContract,
	teleportContract *teleportcontract.TeleportContract,
) Alert {
	return &AlertPegoutFeeNotReset{
		bitcoinClient:         bitcoinClient,
		bitcoinClientContract: bitcoinClientContract,
		teleportContract:      teleportContract,
	}
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

	pegoutBlockHeight := int64(0)
	lastConfirmedBlockHeight := int64(0)
	nextSvb := int64(0)

	// Get info from bitcoin network
	{
		blockHash, err := alert.bitcoinClient.GetBlockHashByTxID(mutils.BytesToBTCHash(pegout.BitcoinTxId))
		if err != nil {
			return SEVERITY_UNKNOWN, labels, err
		}

		blockHeight, err := alert.bitcoinClient.GetBlockHeightByHash(blockHash)
		if err != nil {
			return SEVERITY_UNKNOWN, labels, err
		}

		pegoutBlockHeight = blockHeight
	}

	// Get last confirmed block
	{
		lastConfirmedBlockHash, err := alert.bitcoinClientContract.GetLastConfirmedBlockHash()
		if err != nil {
			return SEVERITY_UNKNOWN, labels, err
		}

		blockHeight, err := alert.bitcoinClient.GetBlockHeightByHash(lastConfirmedBlockHash)
		if err != nil {
			return SEVERITY_UNKNOWN, labels, err
		}

		lastConfirmedBlockHeight = blockHeight
	}

	// Get next_svb
	{
		storage, err := alert.teleportContract.GetStorage(nil)
		if err != nil {
			return SEVERITY_UNKNOWN, labels, err
		}

		nextSvb = int64(storage.NextSVB)
	}

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
