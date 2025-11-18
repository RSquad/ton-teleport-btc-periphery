package ton

import (
	"encoding/hex"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/xssnick/tonutils-go/tlb"
)

func (ef *RawEventFilter) logFilterError(tx *tlb.Transaction, err error) {
	logger.Log.Error().
		Str("component", "RawEventFilter").
		Str("tx_hash", hex.EncodeToString(tx.Hash)).
		Err(err).
		Msg("Error extracting external outs from txs")
}

func (ef *RawEventFilter) logStartWork() {
	logger.Log.Info().
		Str("component", "RawEventFilter").
		Msg("started")
}

func (ef *RawEventFilter) logFinishWork(err error) {
	logger.Log.Info().Err(err).
		Str("component", "RawEventFilter").
		Msg("finished")
}
