package mintservice

import (
	"fmt"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
)

var (
	errQueryUnprocessedMints            = "failed to query unprocessed mints: %w"
	errCalcBitcoinTxID                  = "failed to calculate bitcoin transaction id: %w"
	errCreatePeginContractFromStateInit = "failed to create pegin contract from state init: %w"
	errGetPeginContractAccountState     = "failed to get pegin contract account state: %w"
	errQueryInternalKey                 = "failed to query internal key: %w"
	errQueryLatestInternalKey           = "failed to query latest internal key: %w"
)

func (c *MintService) logStartWork() {
	logger.Log.Info().
		Str("component", "MintService").
		Msg("Start processing mints")
}

func (c *MintService) logFinishWork(err error) {
	event := logger.Log.Info().
		Str("component", "MintService")

	if err != nil {
		event.Err(err).Msg("Finished processing mints with error")
	} else {
		event.Msg("Finished processing mints")
	}
}

func logStartProcessingMints() {
	logger.Log.Info().
		Str("component", "MintService").
		Msg("Start processing mints")
}

func logFinishProcessingMints(duration time.Duration, err error) {
	event := logger.Log.Info().
		Str("component", "MintService")

	if err != nil {
		event.Err(err).Msg("Finished processing mints with error")
	} else {
		event.Dur("duration", duration).
			Msg("Finished processing mints")
	}
}

func logUnprocessedMintsReceived(count int) {
	logger.Log.Info().
		Str("component", "MintService").
		Int("count", count).
		Msg("Unprocessed mints received")
}

func logNoUnprocessedMints() {
	logger.Log.Info().
		Str("component", "MintService").
		Msg("No unprocessed mints")
}

func logMintsProcessingProgress(count int, total int, intervals ...int) {
	interval := 48
	if len(intervals) > 0 {
		interval = intervals[0]
	}

	if count%interval == 0 || count == total {
		logger.Log.Info().
			Str("component", "MintService").
			Str("progress", fmt.Sprintf("%d/%d", count, total)).
			Msg("Mints processing progress")
	}
}
