package alerts

import (
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
)

func logDkgFetchError(err error) {
	logger.Log.Error().
		Str("component", "AlertDkgStatus").
		Err(err).
		Msg("Failed to fetch DKG status")
}

func logNoDkgFound() {
	logger.Log.Debug().
		Str("component", "AlertDkgStatus").
		Msg("No DKG found in database")
}
