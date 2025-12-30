package alerts

import (
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
)

func logDkgStartFetchError(err error) {
	logger.Log.Error().
		Str("component", "AlertDkgRestarts").
		Err(err).
		Msg("Failed to fetch last DKG start event")
}

func logNoDkgStartEvents() {
	logger.Log.Debug().
		Str("component", "AlertDkgRestarts").
		Msg("No DKG start events found")
}

func logDkgRestartsFetchError(txLT uint64, err error) {
	logger.Log.Error().
		Str("component", "AlertDkgRestarts").
		Uint64("start_tx_lt", txLT).
		Err(err).
		Msg("Failed to fetch DKG restart events")
}

func logInfoProcessingError(err error) {
	logger.Log.Error().
		Str("component", "AlertDkgRestarts").
		Err(err).
		Msg("Failed to process DKG info")
}

func logDkgRestartsAlert(severity Severity, info *info) {
	logger.Log.Warn().
		Str("component", "AlertDkgRestarts").
		Str("severity", severity.String()).
		Int("restarts_count", info.restartsCount).
		Int("culprits_count", info.culpritsCount).
		Int("timeout_evicted_count", info.timeoutEvictedCount).
		Int("participants_count", info.participantsCount).
		Int("evicted_ids_count", len(info.evictedIds)).
		Int("all_evicted_ids_count", len(info.allEvictedIds)).
		Msg("DKG restarts alert triggered")
}

func logDkgCheckPassed(info *info) {
	logger.Log.Debug().
		Str("component", "AlertDkgRestarts").
		Int("restarts_count", info.restartsCount).
		Int("culprits_count", info.culpritsCount).
		Int("participants_count", info.participantsCount).
		Msg("DKG restarts check passed")
}
