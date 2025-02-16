package tonclient

import (
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
)

func (tc *TxCollector) logStartWork() {
	logger.Log.Info().
		Str("component", "TxCollector").
		Str("addr", utils.AddrToRawString(tc.addr)).
		Msg("Start collecting txs")
}

func (tc *TxCollector) logFinishWork(err error) {
	event := logger.Log.Info().
		Str("component", "TxCollector").
		Str("addr", utils.AddrToRawString(tc.addr))

	if err != nil {
		event.Err(err).Msg("Finished collecting txs with error")
	} else {
		event.Msg("Finished collecting txs")
	}
}
