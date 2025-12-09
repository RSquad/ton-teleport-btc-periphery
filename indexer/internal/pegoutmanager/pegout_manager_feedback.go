package pegoutmanager

import (
	"fmt"
	"time"

	entpegout "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/pegout"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
)

var (
	errGetPegoutState        = "failed to get pegout state: %w"
	errQuerySigningPegouts   = "failed to query signing pegouts: %w"
	errQuerySignedPegouts    = "failed to query signed pegouts: %w"
	errGetPegoutTxParts      = "failed to get pegout tx parts: %w"
	errBuildPegoutTx         = "failed to build pegout tx: %w"
	errSerializePegoutTx     = "failed to serialize pegout tx: %w"
	errUpdatePegoutToSigned  = "failed to update pegout to signed: %w"
	errParseTxHash           = "failed to parse tx hash: %w"
	errUpdatePegoutToConfirm = "failed to update pegout to confirmed: %w"
	errDeserializePegoutTx   = "failed to deserialize pegout tx for send: %w"
	errSendPegoutTx          = "failed to send pegout tx: %w"
)

func (pm *PegoutManager) logStartManager() {
	logger.Log.Info().
		Str("component", "PegoutManager").
		Msg("PegoutManager started")
}

func (pm *PegoutManager) logStopManager(err error) {
	event := logger.Log.Info().
		Str("component", "PegoutManager")
	if err != nil {
		event.Err(err).Msg("PegoutManager stopped with error")
	} else {
		event.Msg("PegoutManager stopped")
	}
}

func logContextCancelled() {
	logger.Log.Info().Str("component", "PegoutManager").Msg("Context cancelled, waiting for workers to finish...")
}

func (pm *PegoutManager) logStartSigningWork() {
	logger.Log.Info().
		Str("component", "PegoutManager").
		Str("process", "SigningPegouts").
		Msg("Start worker for signing pegouts")
}

func (pm *PegoutManager) logFinishSigningWork(err error) {
	event := logger.Log.Info().
		Str("component", "PegoutManager").
		Str("process", "SigningPegouts")
	if err != nil {
		event.Err(err).Msg("Signing pegouts worker finished with error")
	} else {
		event.Msg("Signing pegouts worker finished")
	}
}

func logStartProcessingSigningPegouts() {
	logger.Log.Debug().
		Str("component", "PegoutManager").
		Str("process", "SigningPegouts").
		Msg("Start processing cycle for signing pegouts")
}

func logFinishProcessingSigningPegouts(duration time.Duration, err error, count int) {
	event := logger.Log.Debug().
		Str("component", "PegoutManager").
		Str("process", "SigningPegouts").
		Dur("duration", duration).
		Int("processed_in_cycle", count)

	if err != nil {
		event.Err(err).Msg("Finished processing cycle for signing pegouts with error")
	} else {
		event.Msg("Finished processing cycle for signing pegouts")
	}
}

func logSigningPegoutsReceived(count int) {
	logger.Log.Info().
		Str("component", "PegoutManager").
		Str("process", "SigningPegouts").
		Int("count", count).
		Msg("Signing pegouts received for processing")
}

func logNoSigningPegouts() {
	logger.Log.Debug().
		Str("component", "PegoutManager").
		Str("process", "SigningPegouts").
		Msg("No signing pegouts to process in this cycle")
}

func logSigningPegoutProgress(count int, total int, intervals ...int) {
	interval := 48
	if len(intervals) > 0 {
		interval = intervals[0]
	}
	if count%interval == 0 || count == total {
		logger.Log.Info().
			Str("component", "PegoutManager").
			Str("process", "SigningPegouts").
			Str("progress", fmt.Sprintf("%d/%d", count, total)).
			Msg("Signing pegouts processing progress")
	}
}

func logFailedProcessSigningPegout(err error, addr string) {
	logger.Log.Error().
		Err(err).
		Str("component", "PegoutManager").
		Str("process", "SigningPegouts").
		Str("pegout_addr", addr).
		Msg("Failed to process signing pegout")
}

func (pm *PegoutManager) logStartSignedWork() {
	logger.Log.Info().
		Str("component", "PegoutManager").
		Str("process", "SignedPegouts").
		Msg("Start worker for signed pegouts")
}

func (pm *PegoutManager) logFinishSignedWork(err error) {
	event := logger.Log.Info().
		Str("component", "PegoutManager").
		Str("process", "SignedPegouts")
	if err != nil {
		event.Err(err).Msg("Signed pegouts worker finished with error")
	} else {
		event.Msg("Signed pegouts worker finished")
	}
}

func logStartProcessingSignedPegouts() {
	logger.Log.Debug().
		Str("component", "PegoutManager").
		Str("process", "SignedPegouts").
		Msg("Start processing cycle for signed pegouts")
}

func logFinishProcessingSignedPegouts(duration time.Duration, err error, processedCount int, excludedInCycleCount int, totalInMemoryExclusions int) {
	event := logger.Log.Debug().
		Str("component", "PegoutManager").
		Str("process", "SignedPegouts").
		Dur("duration", duration).
		Int("processed_in_cycle", processedCount).
		Int("excluded_from_cycle", excludedInCycleCount).
		Int("total_in_memory_exclusions", totalInMemoryExclusions)

	if err != nil {
		event.Err(err).Msg("Finished processing cycle for signed pegouts with error")
	} else {
		event.Msg("Finished processing cycle for signed pegouts")
	}
}

func logSignedPegoutsReceived(count int) {
	logger.Log.Info().
		Str("component", "PegoutManager").
		Str("process", "SignedPegouts").
		Int("count", count).
		Msg("Signed pegouts received for processing")
}

func logNoSignedPegouts() {
	logger.Log.Debug().
		Str("component", "PegoutManager").
		Str("process", "SignedPegouts").
		Msg("No signed pegouts to process in this cycle")
}

func logSignedPegoutProgress(count int, total int, intervals ...int) {
	interval := 48
	if len(intervals) > 0 {
		interval = intervals[0]
	}
	if count%interval == 0 || count == total {
		logger.Log.Info().
			Str("component", "PegoutManager").
			Str("process", "SignedPegouts").
			Str("progress", fmt.Sprintf("%d/%d", count, total)).
			Msg("Signed pegouts processing progress")
	}
}

func logFailedProcessSignedPegout(err error, addr string, txID string) {
	logger.Log.Error().
		Err(err).
		Str("component", "PegoutManager").
		Str("process", "SignedPegouts").
		Str("pegout_addr", addr).
		Str("bitcoin_tx_id", txID).
		Msg("Failed to process signed pegout")
}

func logPegoutTxSent(addr string, txID string) {
	logger.Log.Info().
		Str("component", "PegoutManager").
		Str("process", "SignedPegouts").
		Str("pegout_addr", addr).
		Str("bitcoin_tx_id", txID).
		Msg("Pegout bitcoin transaction sent")
}

func logFailedUpdatePegoutStatus(err error, pegoutID int, newStatus entpegout.Status, bitcoinTxID ...string) {
	event := logger.Log.Error().
		Err(err).
		Str("component", "PegoutManager").
		Int("pegout_id", pegoutID).
		Str("target_status", newStatus.String())
	if len(bitcoinTxID) > 0 {
		event.Str("bitcoin_tx_id", bitcoinTxID[0])
	}
	event.Msg("Failed to update pegout status")
}

func logPegoutStatusUpdated(pegoutID int, newStatus entpegout.Status, bitcoinTxID ...string) {
	event := logger.Log.Info().
		Str("component", "PegoutManager").
		Int("pegout_id", pegoutID).
		Str("new_status", newStatus.String())
	if len(bitcoinTxID) > 0 {
		event.Str("bitcoin_tx_id", bitcoinTxID[0])
	}
	event.Msg("Pegout status updated")
}

func logSigningCycleError(err error) {
	logger.Log.Error().Err(err).Str("component", "PegoutManager").Str("process", "SigningPegouts").Msg("Error during signing pegouts cycle")
}

func logSignedCycleError(err error) {
	logger.Log.Error().Err(err).Str("component", "PegoutManager").Str("process", "SignedPegouts").Msg("Error during signed pegouts cycle")
}

func logSignedPegoutExcluded(pegoutID int, btcTxID string, reasonError string) {
	logger.Log.Warn().
		Str("component", "PegoutManager").
		Str("process", "SignedPegouts").
		Int("pegout_id", pegoutID).
		Str("bitcoin_tx_id", btcTxID).
		Str("reason", reasonError).
		Msg("Signed pegout excluded from further processing in this session due to critical error and age")
}

func logPegoutContractNotActive(addr string) {
	logger.Log.Warn().
		Str("component", "PegoutManager").
		Str("pegout_addr", addr).
		Msg("Pegout contract is not active, skipping.")
}

func logCouldNotDeterminePegoutAge(pegoutID int, btcTxID string, queryError error) {
	logger.Log.Warn().
		Err(queryError).
		Str("component", "PegoutManager").
		Str("process", "SignedPegouts").
		Int("pegout_id", pegoutID).
		Str("bitcoin_tx_id", btcTxID).
		Msg("Could not determine pegout age for exclusion; age check skipped for this attempt.")
}

func logFailedGetMasterchainInfo(err error) {
	logger.Log.Error().
		Err(err).
		Str("component", "PegoutManager").
		Str("process", "SignedPegouts").
		Msg("Failed to get masterchain info")
}
