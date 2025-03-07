package pegoutsigner

import (
	"github.com/rs/zerolog"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
)

func infoEvent() *zerolog.Event {
	return logger.Log.Info().Str("component", "SignService")
}

func errorEvent() *zerolog.Event {
	return logger.Log.Error().Str("component", "SignService")
}

func (s *SignService) logMessage(msg string) {
	infoEvent().Msg(msg)
}

func (s *SignService) logError(err error) {
	errorEvent().Msgf("error: %e", err)
}

func (s *SignService) logCommit(id uint64) {
	infoEvent().Msgf("Commit pegout %x", id)
}
