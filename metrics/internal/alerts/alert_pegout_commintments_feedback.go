package alerts

import (
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
)

func logUnsignedPegoutFetchError(component string, err error) {
	logger.Log.Error().
		Str("component", component).
		Err(err).
		Msg("Failed to fetch first unsigned pegout")
}

func logNoUnsignedPegouts(component string) {
	logger.Log.Debug().
		Str("component", component).
		Msg("No unsigned pegouts found")
}

func logPegoutCommitmentsFetchError(addr *address.Address, err error) {
	logger.Log.Error().
		Str("component", "AlertPegoutCommitments").
		Str("pegout_address", utils.AddrToRawString(addr)).
		Err(err).
		Msg("Failed to fetch pegout details")
}

func logPegoutNotFound(component string, addr *address.Address, err error) {
	logger.Log.Error().
		Str("component", component).
		Str("pegout_address", utils.AddrToRawString(addr)).
		Err(err).
		Msg("Pegout not found in database")
}

func logSigningNotStarted(addr *address.Address) {
	logger.Log.Debug().
		Str("component", "AlertPegoutCommitments").
		Str("pegout_address", utils.AddrToRawString(addr)).
		Msg("Signing stage not started yet for pegout")
}

func logPrevDkgFetchError(component string, err error) {
	logger.Log.Error().
		Str("component", component).
		Err(err).
		Msg("Failed to fetch previous DKG")
}
