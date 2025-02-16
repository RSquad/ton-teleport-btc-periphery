package tonclient

import (
	"encoding/hex"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
)

func (tf *TxFetcher) logStartWork() {
	logger.Log.Info().
		Str("component", "TxFetcher").
		Str("addr", utils.AddrToRawString(tf.addr)).
		Str("start_tx_hash", hex.EncodeToString(tf.hash)).
		Uint64("start_tx_lt", tf.lt).
		Msg("Start fetching txs")
}

func (tf *TxFetcher) logFinishWork(count int, duration time.Duration, err error) {
	event := logger.Log.Info().
		Str("component", "TxFetcher").
		Str("addr", utils.AddrToRawString(tf.addr)).
		Int("count", count).
		Dur("duration", duration)

	if err != nil {
		event.Err(err).Msg("Finished fetching txs with error")
	} else {
		event.Msg("Finished fetching txs")
	}
}

func (tf *TxFetcher) logFetchError(err error) {
	logger.Log.Error().
		Str("component", "TxFetcher").
		Str("addr", utils.AddrToRawString(tf.addr)).
		Err(err).
		Msg("Error fetching txs")
}

func (tf *TxFetcher) logTxsFetched(count int, cumulativeCount int) {
	logger.Log.Info().
		Str("component", "TxFetcher").
		Str("addr", utils.AddrToRawString(tf.addr)).
		Int("count", count).
		Int("cumulative_count", cumulativeCount).
		Msg("Fetched txs")
}
