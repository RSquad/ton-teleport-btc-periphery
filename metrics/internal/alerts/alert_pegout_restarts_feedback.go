package alerts

import (
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
	"github.com/xssnick/tonutils-go/address"
)

func logPegoutRestartsFetchError(component string, addr *address.Address, err error) {
	logger.Log.Error().
		Str("component", component).
		Str("pegout_address", utils.AddrToRawString(addr)).
		Err(err).
		Msg("Failed to fetch pegout details")
}

func logNewPegoutSelected(component string, addr *address.Address, pegout *data_models.Pegout) {
	if pegout == nil {
		return
	}

	logger.Log.Info().
		Str("component", component).
		Str("pegout_address", utils.AddrToRawString(addr)).
		Uint64("pegout_id", pegout.Id).
		Msg("Selected new pegout for monitoring")
}

func logPegoutRestartDetected(component string, addr *address.Address, restartCount int,
	oldExpiredAt, newExpiredAt time.Time,
) {
	logger.Log.Warn().
		Str("component", component).
		Str("pegout_address", utils.AddrToRawString(addr)).
		Int("restart_count", restartCount).
		Time("old_expired_at", oldExpiredAt).
		Time("new_expired_at", newExpiredAt).
		Msg("Pegout restart detected")
}

func logPegoutRestartsAlert(component string, severity Severity, addr *address.Address,
	btcTxId string, restartCount int,
) {
	logger.Log.Warn().
		Str("component", component).
		Str("severity", severity.String()).
		Str("pegout_address", utils.AddrToRawString(addr)).
		Str("bitcoin_tx_id", btcTxId).
		Int("restart_count", restartCount).
		Msg("Pegout restarts alert triggered")
}

func logPegoutCheckPassed(component string, addr *address.Address, btcTxId string, restartCount int) {
	logger.Log.Debug().
		Str("component", component).
		Str("pegout_address", utils.AddrToRawString(addr)).
		Str("bitcoin_tx_id", btcTxId).
		Int("restart_count", restartCount).
		Msg("Pegout restarts check passed")
}
