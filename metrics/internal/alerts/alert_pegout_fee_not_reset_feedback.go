package alerts

import (
	"encoding/hex"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
	"github.com/xssnick/tonutils-go/address"
)

func logLastSignedPegoutFetchError(err error) {
	logger.Log.Error().
		Str("component", "AlertPegoutFeeNotReset").
		Err(err).
		Msg("Failed to fetch last signed pegout")
}

func logNoSignedPegouts() {
	logger.Log.Debug().
		Str("component", "AlertPegoutFeeNotReset").
		Msg("No signed pegouts found")
}

func logNoBitcoinBlockHash(pegout *data_models.Pegout) {
	if pegout == nil {
		return
	}

	pegoutAddr := ""
	if pegout.Addr != nil {
		pegoutAddr = utils.AddrToRawString((*address.Address)(pegout.Addr))
	}

	logger.Log.Debug().
		Str("component", "AlertPegoutFeeNotReset").
		Str("pegout_address", pegoutAddr).
		Uint64("pegout_id", pegout.Id).
		Msg("Pegout has no Bitcoin block hash")
}

func logBlockHeightFetchError(pegout *data_models.Pegout, err error) {
	if pegout == nil {
		return
	}

	pegoutAddr := ""
	if pegout.Addr != nil {
		pegoutAddr = utils.AddrToRawString((*address.Address)(pegout.Addr))
	}

	blockHashHex := ""
	if pegout.BitcoinBlockHash != nil {
		blockHashHex = hex.EncodeToString(pegout.BitcoinBlockHash)
	}

	logger.Log.Error().
		Str("component", "AlertPegoutFeeNotReset").
		Str("pegout_address", pegoutAddr).
		Uint64("pegout_id", pegout.Id).
		Str("bitcoin_block_hash", blockHashHex).
		Err(err).
		Msg("Failed to fetch Bitcoin block height by hash")
}

func logBitcoinClientStorageFetchError(err error) {
	logger.Log.Error().
		Str("component", "AlertPegoutFeeNotReset").
		Err(err).
		Msg("Failed to fetch BitcoinClient contract storage")
}

func logTeleportStorageFetchError(сomponent string, err error) {
	logger.Log.Error().
		Str("component", сomponent).
		Err(err).
		Msg("Failed to fetch Teleport contract storage")
}
