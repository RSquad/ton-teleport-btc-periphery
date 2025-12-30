package fetchers

import (
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
)

func logDataSent(component string, contractName string) {
	logger.Log.Debug().
		Str("component", component).
		Str("contract_name", contractName).
		Msg("Data sent to channel")
}

func logStorageFetchError(component string, contractName string, err error) {
	logger.Log.Error().
		Str("component", component).
		Str("contract_name", contractName).
		Err(err).
		Msg("Failed to retrieve contract storage")
}

func logFetchSuccess(component string, contractName string) {
	logger.Log.Debug().
		Str("component", component).
		Str("contract_name", contractName).
		Msg("Successfully fetched contract data")
}

func logSerializationError(component string, contractName string, err error) {
	logger.Log.Error().
		Str("component", component).
		Err(err).
		Str("contract_name", contractName).
		Msg("Failed to serialize contract data")
}
