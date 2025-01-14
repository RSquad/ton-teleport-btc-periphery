package ton

import (
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/xssnick/tonutils-go/tlb"
)

func (ef *RawEventFilter) logFilterError(tx *tlb.Transaction, err error) {
	logger.Log.Error().
		Str("component", "RawEventFilter").
		Bytes("tx_hash", tx.Hash).
		Err(err).
		Msg("Error extracting external outs from txs")
}
