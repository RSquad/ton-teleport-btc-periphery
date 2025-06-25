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

func (e *Executor) logMessage(dkg *coordinator.DKG, msg string) {
	infoEventWithDkg(dkg, e.validatorIdx).Msg(msg)
}

func (e *Executor) logError(dkg *coordinator.DKG, msg string, err error) {
	errorEventWithDkg(dkg, e.validatorIdx).Err(err).Msg(msg)
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

func (e *Executor) logDKGR1Completed(dkg *coordinator.DKG) {
	if dkg.R1.Count == uint64(dkg.MaxSigners) {
		e.logMessage(dkg, fmt.Sprintf("R1 completed (ready %d)", dkg.R1.Count))
	} else {
		e.logError(dkg, fmt.Sprintf("R1 completed, but MaxSigners(%d) != R1.Count(%d)", dkg.MaxSigners, dkg.R1.Count), nil)
	}
}

func (e *Executor) logDKGR2Completed(dkg *coordinator.DKG) {
	if dkg.R2.Count == uint64(dkg.MaxSigners) {
		e.logMessage(dkg, fmt.Sprintf("R2 completed (ready %d)", dkg.R2.Count))
	} else {
		e.logError(dkg, fmt.Sprintf("R2 completed, but MaxSigners(%d) != R2.Count(%d)", dkg.MaxSigners, dkg.R2.Count), nil)
	}
}

func (e *Executor) logDKGR3Completed(dkg *coordinator.DKG) {
	if dkg.R3.Count == dkg.MaxSigners {
		e.logMessage(dkg, fmt.Sprintf("R3 completed (ready %d)", dkg.R3.Count))
	} else {
		e.logError(dkg, fmt.Sprintf("R3 completed, but MaxSigners(%d) != R3.Count(%d)", dkg.MaxSigners, dkg.R3.Count), nil)
	}
}

func (e *Executor) logSendRound1Package(dkg *coordinator.DKG, err error) {
	errCode, _ := helpers.ExtractExitCode(err.Error())

	if errCode == helpers.ErrDkgExpired {
		infoEventWithDkg(dkg, e.validatorIdx).Msg("DKG Expired")
	} else if errCode == helpers.ErrPackageAlreadyExist {
		infoEventWithDkg(dkg, e.validatorIdx).Msg("Package already exist")
	} else if errCode == helpers.ErrRound1Completed {
		infoEventWithDkg(dkg, e.validatorIdx).Msg("DKG Round1 completed")
	} else {
		msg := helpers.HandleTvmError(err)
		errorEventWithDkg(dkg, e.validatorIdx).Err(err).Msg("failed to send round1 package: " + msg)
	}
}

func (e *Executor) logSendRound2Package(dkg *coordinator.DKG, err error) {
	errCode, _ := helpers.ExtractExitCode(err.Error())

	if errCode == helpers.ErrDkgExpired {
		infoEventWithDkg(dkg, e.validatorIdx).Msg("DKG Expired")
	} else if errCode == helpers.ErrPackageAlreadyExist {
		infoEventWithDkg(dkg, e.validatorIdx).Msg("Package already exist")
	} else {
		msg := helpers.HandleTvmError(err)
		errorEventWithDkg(dkg, e.validatorIdx).Msg("R2 packages sent with errors: " + msg)
	}
}

func (e *Executor) logSendClaimFailed(dkg *coordinator.DKG, culpritIdx uint16, err error) {
	msg := helpers.HandleTvmError(err)

	errorEventWithDkg(dkg, e.validatorIdx).
		Str("culprit validator idx: ", fmt.Sprintf("%d", culpritIdx)).
		Msg("failed to send claim package: " + msg)
}

func (e *Executor) logSendPubkeyPackageFailed(dkg *coordinator.DKG, err error) {
	errCode, _ := helpers.ExtractExitCode(err.Error())

	if errCode == helpers.ErrDkgExpired {
		infoEventWithDkg(dkg, e.validatorIdx).Msg("DKG Expired")
	} else if errCode == helpers.ErrPackageAlreadyExist {
		infoEventWithDkg(dkg, e.validatorIdx).Msg("Package already exist")
	} else {
		msg := helpers.HandleTvmError(err)
		errorEventWithDkg(dkg, e.validatorIdx).Msg("failed to send pubkey package: " + msg)
	}
}
