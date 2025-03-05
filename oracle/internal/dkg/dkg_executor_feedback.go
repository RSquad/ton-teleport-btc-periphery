package dkg

import (
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinatorcontract"
)

func (e *Executor) logStartExecuting(dkg *coordinatorcontract.DKG) {
	logger.Log.Info().
		Str("component", "DKGExecutor").
		Str("dkg_status", dkg.Status.String()).
		Str("dkg_until", dkg.Until.Format(time.RFC3339)).
		Msg("Start executing DKG")
}

func (e *Executor) logFinishExecuting(dkg *coordinatorcontract.DKG) {
	logger.Log.Info().
		Str("component", "DKGExecutor").
		Str("dkg_status", dkg.Status.String()).
		Str("dkg_until", dkg.Until.Format(time.RFC3339)).
		Msg("Finish executing DKG")
}

func (e *Executor) logDKGFinished(dkg *coordinatorcontract.DKG) {
	logger.Log.Info().
		Str("component", "DKGExecutor").
		Str("dkg_status", dkg.Status.String()).
		Str("dkg_until", dkg.Until.Format(time.RFC3339)).
		Msg("DKG finished")
}

func (e *Executor) logNewDKGStarted(dkg *coordinatorcontract.DKG) {
	logger.Log.Info().
		Str("component", "DKGExecutor").
		Str("dkg_status", dkg.Status.String()).
		Str("dkg_until", dkg.Until.Format(time.RFC3339)).
		Msg("New DKG started")
}

func (e *Executor) logDKGProcess(dkg *coordinatorcontract.DKG, msg string) {
	logger.Log.Info().
		Str("component", "DKGExecutor").
		Str("dkg_status", dkg.Status.String()).
		Str("dkg_until", dkg.Until.Format(time.RFC3339)).
		Msg(msg)
}
