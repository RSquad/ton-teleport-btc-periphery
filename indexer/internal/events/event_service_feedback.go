package events

import (
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
)

func (ec *EventService) logStartWork() {
	logger.Log.Info().
		Str("component", "EventService").
		Msg("Start working")
}

func (ec *EventService) logFinishWork(err error) {
	if err != nil {
		logger.Log.Error().
			Str("component", "EventService").
			Err(err).
			Msg("Finished work with error")
	} else {
		logger.Log.Info().
			Str("component", "EventService").
			Msg("Finished work")
	}
}
