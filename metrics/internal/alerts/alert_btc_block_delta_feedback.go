package alerts

import (
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
)

func logStorageReadError(err error) {
	logger.Log.Error().
		Str("component", "AlertBtcBlockDelta").
		Err(err).
		Msg("Failed to read BitcoinClient contract storage")
}

func logBtcHeightFetchError(err error) {
	logger.Log.Error().
		Str("component", "AlertBtcBlockDelta").
		Err(err).
		Msg("Failed to fetch Bitcoin network block height")
}

func logAlertTriggered(severity Severity, delta int,
	contractHeight int, lastConfirmed int64, confirmationsNeeded int64, btcHeight int,
) {
	logger.Log.Warn().
		Str("component", "AlertBtcBlockDelta").
		Str("severity", severity.String()).
		Int("delta", delta).
		Int("contract_height", contractHeight).
		Int64("last_confirmed", lastConfirmed).
		Int64("confirmations_needed", confirmationsNeeded).
		Int("btc_height", btcHeight).
		Msg("BTC block height delta alert triggered")
}

func logAlertCheckPassed(delta int, contractHeight, btcHeight int) {
	logger.Log.Debug().
		Str("component", "AlertBtcBlockDelta").
		Int("delta", delta).
		Int("contract_height", contractHeight).
		Int("btc_height", btcHeight).
		Msg("BTC block height delta check passed")
}

func logBtcHeightRetry(tryId int, err error) {
	logger.Log.Warn().
		Str("component", "AlertBtcBlockDelta").
		Int("try_id", tryId).
		Err(err).
		Msg("Retrying BTC height fetch")
}

func logBtcHeightMaxRetriesExceeded(err error) {
	logger.Log.Error().
		Str("component", "AlertBtcBlockDelta").
		Err(err).
		Msg("Max retries exceeded for BTC height fetch")
}
