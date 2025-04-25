package dkg

import (
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	helpers "github.com/rsquad/ton-teleport-btc-periphery/oracle/internal"
)

const component = "DKGExecutor"

func infoEvent() *zerolog.Event {
	return logger.Log.Info().Str("component", component)
}

func errorEvent() *zerolog.Event {
	return logger.Log.Error().Str("component", component)
}

func infoEventWithDkg(dkg *coordinator.DKG, validatorIdx uint16) *zerolog.Event {
	validatorIdxStr := "unknown"
	if validatorIdx < 255 {
		validatorIdxStr = fmt.Sprintf("%d", validatorIdx)
	}

	return infoEvent().
		Str("validator_idx", validatorIdxStr).
		Str("dkg_state", dkg.State.String()).
		Str("dkg_until", dkg.Until.Format(time.RFC3339))
}

func errorEventWithDkg(dkg *coordinator.DKG, validatorIdx uint16) *zerolog.Event {
	validatorIdxStr := "unknown"
	if validatorIdx < 255 {
		validatorIdxStr = fmt.Sprintf("%d", validatorIdx)
	}

	return errorEvent().
		Str("validator_idx", validatorIdxStr).
		Str("state", dkg.State.String()).
		Str("until", dkg.Until.Format(time.RFC3339))
}

func (e *Executor) logMessage(dkg *coordinator.DKG, validatorIdx uint16, msg string) {
	infoEventWithDkg(dkg, validatorIdx).Msg(msg)
}

func (e *Executor) logError(dkg *coordinator.DKG, validatorIdx uint16, msg string, err error) {
	errorEventWithDkg(dkg, validatorIdx).Err(err).Msg(msg)
}

func (e *Executor) logStartExecuting(dkg *coordinator.DKG, validatorIdx uint16) {
	e.logMessage(dkg, validatorIdx, "start")
}

func (e *Executor) logFinishExecuting(dkg *coordinator.DKG, validatorIdx uint16) {
	e.logMessage(dkg, validatorIdx, "stop")
}

func (e *Executor) logDKGFinished(dkg *coordinator.DKG, validatorIdx uint16) {
	e.logMessage(dkg, validatorIdx, "DKG finished")
}

func (e *Executor) logNewDKGStarted(dkg *coordinator.DKG, validatorIdx uint16) {
	e.logMessage(dkg, validatorIdx, "new DKG started")
}

func (e *Executor) logDKGProcess(dkg *coordinator.DKG, validatorIdx uint16, msg string) {
	e.logMessage(dkg, validatorIdx, msg)
}

func (e *Executor) logDKGPart1Failed(dkg *coordinator.DKG, validatorIdx uint16, err error) {
	e.logError(dkg, validatorIdx, "part1 failed", err)
}

func (e *Executor) logExecuteR1(dkg *coordinator.DKG, validatorIdx uint16) {
	e.logMessage(dkg, validatorIdx, "execute R1")
}

func (e *Executor) logExecuteR2(dkg *coordinator.DKG, validatorIdx uint16) {
	e.logMessage(dkg, validatorIdx, "execute R2")
}

func (e *Executor) logExecuteClaim(dkg *coordinator.DKG, validatorIdx uint16) {
	e.logMessage(dkg, validatorIdx, "execute claim")
}

func (e *Executor) logExecuteR3(dkg *coordinator.DKG, validatorIdx uint16) {
	e.logMessage(dkg, validatorIdx, "execute R3")
}

func (e *Executor) logSendRound1Package(dkg *coordinator.DKG, validatorIdx uint16, err error) {
	msg := helpers.HandleTvmError(err)
	errorEventWithDkg(dkg, validatorIdx).Err(err).Msg("failed to send round1 package: " + msg)
}

func (e *Executor) logSendRound2Package(dkg *coordinator.DKG, validatorIdx uint16, err error) {
	msg := helpers.HandleTvmError(err)
	errorEventWithDkg(dkg, validatorIdx).
		Msg("R2 packages sent with errors: " + msg)
}

func (e *Executor) logSendClaimFailed(dkg *coordinator.DKG, validatorIdx uint16, culpritIdx uint16, err error) {
	msg := helpers.HandleTvmError(err)

	errorEventWithDkg(dkg, validatorIdx).
		Str("culprit validator idx: ", fmt.Sprintf("%d", culpritIdx)).
		Msg("failed to send claim package: " + msg)
}

func (e *Executor) logSendPubkeyPackageFailed(dkg *coordinator.DKG, validatorIdx uint16, err error) {
	msg := helpers.HandleTvmError(err)
	errorEventWithDkg(dkg, validatorIdx).Msg("failed to send pubkey package: " + msg)
}
