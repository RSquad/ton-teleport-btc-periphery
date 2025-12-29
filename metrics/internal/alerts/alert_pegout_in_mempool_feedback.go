package alerts

import (
	"encoding/hex"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
	"github.com/xssnick/tonutils-go/address"
)

func logSignedPegoutsFetchError(component string, err error) {
	logger.Log.Error().
		Str("component", component).
		Err(err).
		Msg("Failed to fetch last signed pegouts")
}

func logNoSignedPegoutsFound(component string) {
	logger.Log.Debug().
		Str("component", component).
		Msg("No signed pegouts found")
}

func logNewPegoutToCheck(component string, pegout *data_models.Pegout) {
	if pegout == nil {
		return
	}

	pegoutAddr := ""
	if pegout.Addr != nil {
		pegoutAddr = utils.AddrToRawString((*address.Address)(pegout.Addr))
	}

	logger.Log.Info().
		Str("component", component).
		Str("pegout_address", pegoutAddr).
		Uint64("pegout_id", pegout.Id).
		Msg("Selected new pegout to check")
}

func logNoNewPegoutsToCheck(component string, lastCheckedId uint64) {
	logger.Log.Debug().
		Str("component", component).
		Uint64("last_checked_id", lastCheckedId).
		Msg("No new pegouts to check")
}

func logInvalidBitcoinTxId(component string, pegout *data_models.Pegout, err error) {
	if pegout == nil {
		return
	}

	pegoutAddr := ""
	if pegout.Addr != nil {
		pegoutAddr = utils.AddrToRawString((*address.Address)(pegout.Addr))
	}

	btcTxIdHex := ""
	if len(pegout.BitcoinTxId) > 0 {
		btcTxIdHex = hex.EncodeToString(pegout.BitcoinTxId)
	}

	logger.Log.Error().
		Str("component", component).
		Str("pegout_address", pegoutAddr).
		Uint64("pegout_id", pegout.Id).
		Str("bitcoin_tx_id", btcTxIdHex).
		Int("tx_id_length", len(pegout.BitcoinTxId)).
		Err(err).
		Msg("Invalid Bitcoin transaction ID")
}

func logFoundInMempool(component string, pegout *data_models.Pegout, btcTxIdHex string) {
	if pegout == nil {
		return
	}

	logger.Log.Debug().
		Str("component", component).
		Str("bitcoin_tx_id", btcTxIdHex).
		Uint64("pegout_id", pegout.Id).
		Msg("Pegout found in Bitcoin mempool")
}

func logMempoolCheckError(component string, pegout *data_models.Pegout, btcTxIdHex string, err error) {
	if pegout == nil {
		return
	}

	logger.Log.Warn().
		Str("component", component).
		Str("bitcoin_tx_id", btcTxIdHex).
		Uint64("pegout_id", pegout.Id).
		Err(err).
		Msg("Error checking Bitcoin mempool")
}

func logFoundInBlock(component string, pegout *data_models.Pegout, btcTxIdHex, blockHash string) {
	if pegout == nil {
		return
	}

	logger.Log.Debug().
		Str("component", component).
		Str("bitcoin_tx_id", btcTxIdHex).
		Str("block_hash", blockHash).
		Uint64("pegout_id", pegout.Id).
		Msg("Pegout found in Bitcoin block")
}

func logBlockCheckError(component string, pegout *data_models.Pegout, btcTxIdHex string, err error) {
	if pegout == nil {
		return
	}

	logger.Log.Warn().
		Str("component", component).
		Str("bitcoin_tx_id", btcTxIdHex).
		Uint64("pegout_id", pegout.Id).
		Err(err).
		Msg("Error checking Bitcoin blocks")
}

func logPegoutMissingDuration(component string, pegout *data_models.Pegout, btcTxIdHex string, duration time.Duration) {
	if pegout == nil {
		return
	}

	minutes := int(duration.Minutes())

	logger.Log.Info().
		Str("component", component).
		Str("bitcoin_tx_id", btcTxIdHex).
		Uint64("pegout_id", pegout.Id).
		Dur("duration", duration).
		Int("duration_minutes", minutes).
		Msg("Pegout still missing from Bitcoin network")
}

func logPegoutMissingAlert(component string, severity Severity, pegout *data_models.Pegout, btcTxIdHex string, duration time.Duration) {
	if pegout == nil {
		return
	}

	pegoutAddr := ""
	if pegout.Addr != nil {
		pegoutAddr = utils.AddrToRawString((*address.Address)(pegout.Addr))
	}

	minutes := int(duration.Minutes())

	logger.Log.Warn().
		Str("component", component).
		Str("severity", severity.String()).
		Str("pegout_address", pegoutAddr).
		Uint64("pegout_id", pegout.Id).
		Str("bitcoin_tx_id", btcTxIdHex).
		Dur("duration", duration).
		Int("duration_minutes", minutes).
		Msg("Pegout missing alert triggered")
}

func logPegoutFoundSuccessfully(component string, btcTxIdHex string) {
	logger.Log.Info().
		Str("component", component).
		Str("bitcoin_tx_id", btcTxIdHex).
		Msg("Pegout found successfully in Bitcoin network")
}
