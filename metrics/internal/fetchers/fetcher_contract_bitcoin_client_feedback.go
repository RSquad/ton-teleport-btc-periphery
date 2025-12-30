package fetchers

import (
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
)

func logStorageCellError(component string, err error) {
	logger.Log.Error().
		Str("component", component).
		Err(err).
		Msg("Failed to retrieve storage cell")
}

func logCandidateBlockHashesError(component string, err error) {
	logger.Log.Error().
		Str("component", component).
		Err(err).
		Msg("Failed to retrieve candidate block hashes")
}

func logLastConfirmedBlockHashError(component string, err error) {
	logger.Log.Error().
		Str("component", component).
		Err(err).
		Msg("Failed to retrieve last confirmed block hash")
}

func logBlockHeightError(component string, blockHash string, err error) {
	logger.Log.Error().
		Str("component", component).
		Str("block_hash", blockHash).
		Err(err).
		Msg("Failed to retrieve block height by hash")
}
