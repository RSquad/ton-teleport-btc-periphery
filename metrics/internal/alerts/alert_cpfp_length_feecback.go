package alerts

import (
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
	"github.com/xssnick/tonutils-go/address"
)

func logPegoutFetchError(component string, err error) {
	logger.Log.Error().
		Str("component", component).
		Err(err).
		Msg("Failed to fetch last confirmed pegout")
}

func logNoPegoutsFound(component string) {
	logger.Log.Debug().
		Str("component", component).
		Msg("No confirmed pegouts found")
}

func logNullBitcoinTxId(pegout *data_models.Pegout) {
	if pegout == nil {
		return
	}

	pegoutAddr := ""
	if pegout.Addr != nil {
		pegoutAddr = utils.AddrToRawString((*address.Address)(pegout.Addr))
	}

	logger.Log.Error().
		Str("component", "AlertCpfpLength").
		Str("pegout_address", pegoutAddr).
		Uint64("pegout_id", pegout.Id).
		Msg("Bitcoin TxId is null for pegout")
}

func logCpfpLengthFetchError(btcTxIdHex string, err error) {
	logger.Log.Error().
		Str("component", "AlertCpfpLength").
		Str("bitcoin_tx_id", btcTxIdHex).
		Err(err).
		Msg("Failed to fetch CPFP chain length")
}
