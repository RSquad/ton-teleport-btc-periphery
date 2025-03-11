package dkg

import (
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
)

func (e *Executor) logStartExecuting(dkg *coordinator.DKG) {
	logger.Log.Info().
		Str("component", "DKGExecutor").
		Str("dkg_status", dkg.Status.String()).
		Str("dkg_until", dkg.Until.Format(time.RFC3339)).
		Msg("start")
}

func (e *Executor) logFinishExecuting(dkg *coordinator.DKG) {
	logger.Log.Info().
		Str("component", "DKGExecutor").
		Str("dkg_status", dkg.Status.String()).
		Str("dkg_until", dkg.Until.Format(time.RFC3339)).
		Msg("stop")
}

func (e *Executor) logDKGFinished(dkg *coordinator.DKG) {
	logger.Log.Info().
		Str("component", "DKGExecutor").
		Str("dkg_status", dkg.Status.String()).
		Str("dkg_until", dkg.Until.Format(time.RFC3339)).
		Msg("DKG finished")
}

func (e *Executor) logNewDKGStarted(dkg *coordinator.DKG) {
	logger.Log.Info().
		Str("component", "DKGExecutor").
		Str("dkg_status", dkg.Status.String()).
		Str("dkg_until", dkg.Until.Format(time.RFC3339)).
		Msg("New DKG started")
}

func (e *Executor) logDKGProcess(dkg *coordinator.DKG, msg string) {
	logger.Log.Info().
		Str("component", "DKGExecutor").
		Str("dkg_status", dkg.Status.String()).
		Str("dkg_until", dkg.Until.Format(time.RFC3339)).
		Msg(msg)
}

func (e *Executor) logDKGPart1Failed(dkg *coordinator.DKG, err error) {
	logger.Log.Error().
		Str("component", "DKGExecutor").
		Str("dkg_status", dkg.Status.String()).
		Str("dkg_until", dkg.Until.Format(time.RFC3339)).
		Msgf("part1 failed: %e", err)
}
