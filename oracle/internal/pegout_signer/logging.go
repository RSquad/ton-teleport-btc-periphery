package pegoutsigner

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
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

func (s *SignService) logMsgf(format string, v ...interface{}) {
	infoEvent().Msgf(format, v)
}

func (s *SignService) logError(msg string, err error) {
	errorEvent().Err(err).Msg(msg)
}

func (s *SignService) logCommitPegout(pegoutID uint64) {
	infoEvent().Msgf("Commit pegout %x", pegoutID)
}

func (s *SignService) logProcessingPegout(pegout *coordinator.PegoutRecord) {
	infoEvent().Msgf("Processing pegout ID: %x", pegout.ID)
	infoEvent().Msgf("Pegout address: %s", pegout.PegoutAddress)
}

func (s *SignService) logOracleNotValidator(pegoutID uint64) {
	err := fmt.Errorf("Oracle is not a validator. Cannot participate in signing pegout: %x", pegoutID)
	errorEvent().Err(err)
}

func (s *SignService) logErrNullNonceOrCommitments(nonce []byte, commitments []byte, pegoutAddrStr string) {
	var err error = nil
	if nonce == nil {
		err = fmt.Errorf("Failed to load nonce for %s", pegoutAddrStr)
	} else if commitments == nil {
		err = fmt.Errorf("Failed to load commitments for %s", pegoutAddrStr)
	}
	errorEvent().Err(err)
}

func (s *SignService) logErrNoOracleCommitments(pegoutID uint64) {
	err := fmt.Errorf("Oracle didn't send commitment and cannot participate in signing for pegout %x", pegoutID)
	errorEvent().Err(err)
}

func (s *SignService) logPegoutSigned(pegoutID uint64) {
	infoEvent().Msgf("Pegout %x signed", pegoutID)
}
