package pegout

import (
	"encoding/hex"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
)

func logPegoutCreateError(err error, event ton.EventInterface) {
	logger.Log.Error().
		Str("component", "PegoutWriter").
		Err(err).
		Str("tx_hash", hex.EncodeToString(event.GetRaw().TxHash)).
		Msg("Pegout write error")
}
