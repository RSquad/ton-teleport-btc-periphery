package alerts

import (
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
)

func logCoordinatorStorageFetchError(component string, err error) {
	logger.Log.Error().
		Str("component", component).
		Err(err).
		Msg("Failed to fetch coordinator contract storage")
}

func logSigningTimeoutLoaded(component string, signingTimeout uint32) {
	logger.Log.Debug().
		Str("component", component).
		Uint32("signing_timeout", signingTimeout).
		Msg("Signing timeout loaded from coordinator")
}

func logInvalidBeginTimestamp(component string, unsignedPegout *coordinator.PegoutRecord,
	expiredAt int64, signingTimeout uint32, beginTimestamp int64, err error,
) {
	if unsignedPegout == nil {
		return
	}

	logger.Log.Error().
		Str("component", component).
		Str("pegout_address", utils.AddrToRawString(unsignedPegout.PegoutAddress)).
		Int64("expired_at", expiredAt).
		Uint32("signing_timeout", signingTimeout).
		Int64("begin_timestamp", beginTimestamp).
		Err(err).
		Msg("Invalid begin timestamp calculated")
}

func logNewPegoutMonitoringStarted(component string, unsignedPegout *coordinator.PegoutRecord,
	expiredAt, beginTimestamp int64, signingTimeout uint32,
) {
	if unsignedPegout == nil {
		return
	}

	logger.Log.Info().
		Str("component", component).
		Str("pegout_address", utils.AddrToRawString(unsignedPegout.PegoutAddress)).
		Int64("expired_at", expiredAt).
		Int64("begin_timestamp", beginTimestamp).
		Uint32("signing_timeout", signingTimeout).
		Msg("Started monitoring new pegout signing duration")
}

func logSigningDurationIncreased(component string, unsignedPegout *coordinator.PegoutRecord,
	duration time.Duration, currentSeverity, newSeverity Severity,
) {
	if unsignedPegout == nil {
		return
	}

	minutes := int(duration.Minutes())

	logger.Log.Warn().
		Str("component", component).
		Str("pegout_address", utils.AddrToRawString(unsignedPegout.PegoutAddress)).
		Dur("duration", duration).
		Int("duration_minutes", minutes).
		Str("current_severity", currentSeverity.String()).
		Str("new_severity", newSeverity.String()).
		Msg("Pegout signing duration increased, severity updated")
}
