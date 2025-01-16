package tonclient

import (
	"encoding/hex"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/tlb"
)

func (ts *TxSubscriber) logStartWork() {
	logger.Log.Info().
		Str("component", "TxSubscriber").
		Str("addr", utils.AddrToRawString(ts.addr)).
		Uint64("start_tx_lt", ts.lt).
		Msg("Start listening to txs")
}

func (ts *TxSubscriber) logFinishWork(err error) {
	event := logger.Log.Info().
		Str("component", "TxSubscriber").
		Str("addr", utils.AddrToRawString(ts.addr))

	if err != nil {
		event.Err(err).Msg("Finished listening to txs with error")
	} else {
		event.Msg("Finished listening to txs")
	}
}

func (ts *TxSubscriber) logTxReceived(tx *tlb.Transaction) {
	logger.Log.Info().
		Str("component", "TxSubscriber").
		Str("address", utils.AddrToRawString(ts.addr)).
		Str("tx_hash", hex.EncodeToString(tx.Hash)).
		Msg("Received transaction")
}
