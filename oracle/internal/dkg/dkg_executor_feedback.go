package dkg

import (
	"encoding/hex"
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

func infoEventWithDkg(dkg *coordinator.DKG) *zerolog.Event {
	return infoEvent().
		Str("dkg_status", dkg.Status.String()).
		Str("dkg_until", dkg.Until.Format(time.RFC3339))
}

func errorEventWithDkg(dkg *coordinator.DKG) *zerolog.Event {
	return errorEvent().
		Str("status", dkg.Status.String()).
		Str("until", dkg.Until.Format(time.RFC3339))
}

func (e *Executor) logMessage(dkg *coordinator.DKG, msg string) {
	infoEventWithDkg(dkg).Msg(msg)
}

func (e *Executor) logError(dkg *coordinator.DKG, msg string, err error) {
	errorEventWithDkg(dkg).Err(err).Msg(msg)
}

func (e *Executor) logStartExecuting(dkg *coordinator.DKG) {
	e.logMessage(dkg, "start")
}

func (e *Executor) logFinishExecuting(dkg *coordinator.DKG) {
	e.logMessage(dkg, "stop")
}

func (e *Executor) logDKGFinished(dkg *coordinator.DKG) {
	e.logMessage(dkg, "DKG finished")
}

func (e *Executor) logNewDKGStarted(dkg *coordinator.DKG) {
	e.logMessage(dkg, "new DKG started")
}

func (e *Executor) logDKGProcess(dkg *coordinator.DKG, msg string) {
	e.logMessage(dkg, msg)
}

func (e *Executor) logDKGPart1Failed(dkg *coordinator.DKG, err error) {
	e.logError(dkg, "part1 failed", err)
}

func (e *Executor) logExecuteR1(dkg *coordinator.DKG) {
	e.logMessage(dkg, "execute R1")
}

func (e *Executor) logExecuteR2(dkg *coordinator.DKG) {
	e.logMessage(dkg, "execute R2")
}

func (e *Executor) logExecuteClaim(dkg *coordinator.DKG) {
	e.logMessage(dkg, "execute claim")
}

func (e *Executor) logExecuteR3(dkg *coordinator.DKG) {
	e.logMessage(dkg, "execute R3")
}

func (e *Executor) logSendRound1Package(dkg *coordinator.DKG, err error) {
	msg := helpers.HandleTvmError(err)
	errorEventWithDkg(dkg).Err(err).Msg("failed to send round1 package: " + msg)
}

func (e *Executor) logSendRound2Package(dkg *coordinator.DKG, identifierTo []byte, err error) {
	msg := helpers.HandleTvmError(err)
	errorEventWithDkg(dkg).
		Str("to", hex.EncodeToString(identifierTo)).
		Msg("failed to send round2 package: " + msg)
}

func (e *Executor) logSendClaimPackage(dkg *coordinator.DKG, maliciousValidatorIdx []byte, err error) {
	msg := helpers.HandleTvmError(err)

	maliciousValidatorIdxStr := "NO"
	if maliciousValidatorIdx != nil {
		maliciousValidatorIdxStr = hex.EncodeToString(maliciousValidatorIdx)
	}

	errorEventWithDkg(dkg).
		Str("malicious validator idx: ", maliciousValidatorIdxStr).
		Msg("failed to send claim package: " + msg)
}

func (e *Executor) logSendPubkeyPackageFailed(dkg *coordinator.DKG, err error) {
	msg := helpers.HandleTvmError(err)
	errorEventWithDkg(dkg).Msg("failed to send pubkey package: " + msg)
}
