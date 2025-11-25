package mintservice

import (
	"fmt"
	"time"

	mintmodel "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/mint"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
)

var (
	errCalcBitcoinTxID                  = "failed to calculate bitcoin transaction id: %w"
	errCreatePeginContractFromStateInit = "failed to create pegin contract from state init: %w"
	errGetPeginContractAccountState     = "failed to get pegin contract account state: %w"
	errQueryInternalKey                 = "failed to query internal key: %w"
	errQueryLatestInternalKey           = "failed to query latest internal key: %w"
	errQueryPendingMints                = "failed to query pending mints: %w"
	errQueryRefundMints                 = "failed to query refund mints: %w"
)

func (ms *MintService) logStartWork() {
	logger.Log.Info().
		Str("component", "MintService").
		Msg("MintService started")
}

func (ms *MintService) logFinishWork(err error) {
	event := logger.Log.Info().
		Str("component", "MintService")

	if err != nil {
		event.Err(err).Msg("MintService stopped with error")
	} else {
		event.Msg("MintService stopped")
	}
}

func (ms *MintService) logStartPendingWork() {
	logger.Log.Info().
		Str("component", "MintService").
		Str("process", "PendingMints").
		Msg("Start worker for pending mints")
}

func (ms *MintService) logFinishPendingWork(err error) {
	event := logger.Log.Info().
		Str("component", "MintService").
		Str("process", "PendingMints")
	if err != nil {
		event.Err(err).Msg("Pending mints worker finished with error")
	} else {
		event.Msg("Pending mints worker finished")
	}
}

func logStartProcessingPendingMints() {
	logger.Log.Debug().
		Str("component", "MintService").
		Str("process", "PendingMints").
		Msg("Start processing cycle for pending mints")
}

func logFinishProcessingPendingMints(duration time.Duration, err error, count int) {
	event := logger.Log.Debug().
		Str("component", "MintService").
		Str("process", "PendingMints").
		Dur("duration", duration).
		Int("processed_in_cycle", count)

	if err != nil {
		event.Err(err).Msg("Finished processing cycle for pending mints with error")
	} else {
		event.Msg("Finished processing cycle for pending mints")
	}
}

func logPendingMintsReceived(count int) {
	logger.Log.Info().
		Str("component", "MintService").
		Str("process", "PendingMints").
		Int("count", count).
		Msg("Pending mints received for processing")
}

func logNoPendingMints() {
	logger.Log.Debug().
		Str("component", "MintService").
		Str("process", "PendingMints").
		Msg("No pending mints to process in this cycle")
}

func logPendingMintsProcessingProgress(count int, total int, intervals ...int) {
	interval := 48
	if len(intervals) > 0 {
		interval = intervals[0]
	}

	if count%interval == 0 || count == total {
		logger.Log.Info().
			Str("component", "MintService").
			Str("process", "PendingMints").
			Str("progress", fmt.Sprintf("%d/%d", count, total)).
			Msg("Pending mints processing progress")
	}
}

func (ms *MintService) logStartRefundWork() {
	logger.Log.Info().
		Str("component", "MintService").
		Str("process", "RefundMints").
		Msg("Start worker for refund mints")
}

func (ms *MintService) logFinishRefundWork(err error) {
	event := logger.Log.Info().
		Str("component", "MintService").
		Str("process", "RefundMints")
	if err != nil {
		event.Err(err).Msg("Refund mints worker finished with error")
	} else {
		event.Msg("Refund mints worker finished")
	}
}

func logStartProcessingRefundMints() {
	logger.Log.Debug().
		Str("component", "MintService").
		Str("process", "RefundMints").
		Msg("Start processing cycle for refund mints")
}

func logFinishProcessingRefundMints(duration time.Duration, err error, count int) {
	event := logger.Log.Debug().
		Str("component", "MintService").
		Str("process", "RefundMints").
		Dur("duration", duration).
		Int("processed_in_cycle", count)

	if err != nil {
		event.Err(err).Msg("Finished processing cycle for refund mints with error")
	} else {
		event.Msg("Finished processing cycle for refund mints")
	}
}

func logRefundMintsReceived(count int) {
	logger.Log.Info().
		Str("component", "MintService").
		Str("process", "RefundMints").
		Int("count", count).
		Msg("Refund mints received for processing")
}

func logNoRefundMints() {
	logger.Log.Debug().
		Str("component", "MintService").
		Str("process", "RefundMints").
		Msg("No refund mints to process in this cycle")
}

func logRefundMintsProcessingProgress(count int, total int, intervals ...int) {
	interval := 48
	if len(intervals) > 0 {
		interval = intervals[0]
	}

	if count%interval == 0 || count == total {
		logger.Log.Info().
			Str("component", "MintService").
			Str("process", "RefundMints").
			Str("progress", fmt.Sprintf("%d/%d", count, total)).
			Msg("Refund mints processing progress")
	}
}

func logContextCancelled() {
	logger.Log.Info().Str("component", "MintService").Msg("Context cancelled, waiting for workers to finish...")
}

func logPendingCycleError(err error) {
	logger.Log.Error().Err(err).Str("component", "MintService").Str("process", "PendingMints").Msg("Error during pending mints cycle")
}

func logRefundCycleError(err error) {
	logger.Log.Error().Err(err).Str("component", "MintService").Str("process", "RefundMints").Msg("Error during refund mints cycle")
}

func logNoInternalKeysFoundWarning() {
	logger.Log.Warn().Str("component", "MintService").Str("process", "PendingMints").Msg("No internal keys found in the database, proceeding with latestInternalKey as nil")
}

func logFailedProcessPendingMint(err error, mintID int) {
	logger.Log.Error().Err(err).Str("component", "MintService").Str("process", "PendingMints").Int("mint_id", mintID).Msg("Failed to process pending mint")
}

func logPendingMintPeginInternalKeyNotFound(err error, mintID int, peginInternalKey string) {
	logger.Log.Error().Err(err).Str("component", "MintService").Int("mint_id", mintID).Str("pegin_internal_key", peginInternalKey).Msg("Internal key associated with pegin not found.")
}

func logFailedProcessRefundMint(err error, mintID int) {
	logger.Log.Error().Err(err).Str("component", "MintService").Str("process", "RefundMints").Int("mint_id", mintID).Msg("Failed to process refund mint")
}

func logRefundMintFailedParseBitcoinTxID(err error, mintID int, bitcoinTxID string) {
	logger.Log.Error().Err(err).Str("component", "MintService").Int("mint_id", mintID).Str("bitcoin_tx_id", bitcoinTxID).Msg("Failed to parse BitcoinTxID for refund mint")
}

func logRefundMintFailedGetTxOut(err error, mintID int) {
	logger.Log.Error().Err(err).Str("component", "MintService").Int("mint_id", mintID).Msg("Failed to GetTxOut for refund mint")
}

func logFailedUpdateMintStatus(err error, mintID int, status mintmodel.Status) {
	logger.Log.Error().Err(err).Str("component", "MintService").Int("mint_id", mintID).Str("target_status", string(status)).Msg("Failed to update mint status")
}

func logMintStatusUpdated(mintID int, status mintmodel.Status) {
	logger.Log.Info().Str("component", "MintService").Int("mint_id", mintID).Str("new_status", string(status)).Msg("Mint status updated")
}

func logFailedProcessUnconfirmedMint(err error, mintID int) {
	logger.Log.Error().Err(err).Str("component", "MintService").Str("process", "ProcessingUnconfirmedMints").Int("mint_id", mintID).Msg("Failed to process unconfirmed mint")
}

func (ms *MintService) logDepositContextCancelled(mintID int, err error) {
	logger.Log.Warn().
		Str("component", "MintService").
		Int("mint_id", mintID).
		Err(err).
		Msg("Deposit monitoring stopped due to context cancellation")
}

func (ms *MintService) logPeginActivationCheckFailed(mintID int, attemptCount int, err error) {
	logger.Log.Error().
		Str("component", "MintService").
		Int("mint_id", mintID).
		Int("attempt_count", attemptCount).
		Err(err).
		Msg("Failed to check if pegin contract is active")
}

func (ms *MintService) logFailedHandlePendingMint(mintID int, attemptCount int, err error) {
	logger.Log.Error().
		Str("component", "MintService").
		Int("mint_id", mintID).
		Int("attempt_count", attemptCount).
		Err(err).
		Msg("Failed to handle pending mint")
}

func (ms *MintService) logPeginActivationTimeout(mintID int) {
	logger.Log.Warn().
		Str("component", "MintService").
		Int("mint_id", mintID).
		Msg("Pegin contract is not active after timeout")
}

func logSendedMessages(count int) {
	logger.Log.Info().Str("component", "MintService").Int("count", count).Msg("Sended messages")
}

func logFailedCastMessageToMessageWithTxHash(err error, idx int) {
	logger.Log.Error().Err(err).Str("component", "MintService").Int("idx", idx).Msg("Failed to cast message to MessageWithTxHash")
}
