package fetchers

import (
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
)

func logDkgFetchError(component string, err error) {
	logger.Log.Error().
		Str("component", component).
		Err(err).
		Msg("Failed to fetch DKG")
}

func logNoDkgFound(component string) {
	logger.Log.Debug().
		Str("component", component).
		Msg("No DKG found (null returned from contract)")
}

func logNoPrevDkgFound(component string) {
	logger.Log.Debug().
		Str("component", component).
		Msg("No previous DKG found (null returned from contract)")
}
