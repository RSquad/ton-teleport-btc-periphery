package pegoutsigner

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
)

func strPegoutID(pegoutID uint64) string {
	return fmt.Sprintf("%x", pegoutID)
}

func infoEvent() *zerolog.Event {
	return logger.Log.Info().Str("component", "SignService")
}

func infoEventWithPegoutID(pegoutID uint64) *zerolog.Event {
	return logger.Log.Info().
		Str("component", "SignService").
		Str("PegoutId", strPegoutID(pegoutID))
}

func errorEventWithPegoutID(pegoutID uint64) *zerolog.Event {
	return logger.Log.Error().
		Str("component", "SignService").
		Str("PegoutId", strPegoutID(pegoutID))
}

func errorEvent() *zerolog.Event {
	return logger.Log.Error().Str("component", "SignService")
}

func (s *SignService) logMessage(msg string) {
	infoEvent().Msg(msg)
}

func (s *SignService) logError(msg string, err error) {
	errorEvent().Err(err).Msg(msg)
}

func (s *SignService) logCommitPegout(pegoutID uint64) {
	infoEventWithPegoutID(pegoutID).Msg("commit")
}

func (s *SignService) logSignPegout(pegoutID uint64) {
	infoEventWithPegoutID(pegoutID).Msg("sign")
}

func (s *SignService) logProcessingPegout(pegout *coordinator.PegoutRecord) {
	infoEventWithPegoutID(pegout.ID).Msgf("address %s", pegout.PegoutAddress)
}

func (s *SignService) logOracleNotValidator(pegoutID uint64) {
	err := fmt.Errorf("oracle is not a validator. Cannot participate in signing pegout: %x", pegoutID)
	errorEvent().Err(err)
}

func (s *SignService) logErrNullNonceOrCommitments(nonce []byte, commitments []byte, pegoutAddrStr string) {
	var err error = nil
	if nonce == nil {
		err = fmt.Errorf("failed to load nonce for %s", pegoutAddrStr)
	} else if commitments == nil {
		err = fmt.Errorf("failed to load commitments for %s", pegoutAddrStr)
	}
	errorEvent().Err(err)
}

func (s *SignService) logErrNoOracleCommitments(pegoutID uint64) {
	err := fmt.Errorf("oracle didn't send commitment and cannot participate in signing for pegout %x", pegoutID)
	errorEvent().Err(err)
}

func (s *SignService) logPegoutSigned(pegoutID uint64) {
	infoEventWithPegoutID(pegoutID).Msg("pegout signed")
}

func (s *SignService) logSigningShareSent(pegoutID uint64) {
	infoEventWithPegoutID(pegoutID).Msg("signing share sent")
}

func (s *SignService) logGetPrevDKGError(err error) {
	errorEvent().Err(err).Msg("failed to get previous DKG")
}

func (s *SignService) logUnsignedPegoutsError(err error) {
	errorEvent().Err(err).Msg("failed to get unsigned pegouts")
}

func (s *SignService) logSigningRequestsCount(count int) {
	infoEvent().Msgf("%d signing requests", count)
}

func (s *SignService) logCommitSent(pegoutID uint64) {
	infoEventWithPegoutID(pegoutID).Msg("Commit sent")
}

func (s *SignService) logMinimalSharesReached(pegoutID uint64) {
	infoEventWithPegoutID(pegoutID).Msg("Minimal required number of signing shares is reached")
}

func (s *SignService) logSigningShareAlreadyExists(pegoutID uint64) {
	infoEventWithPegoutID(pegoutID).Msg("Signing share already exists")
}

func (s *SignService) logErrNothingToSign(pegoutID uint64) {
	errorEventWithPegoutID(pegoutID).Msg("pegout has no signing hashes")
}

func (s *SignService) logAggregateSignShares(pegoutID uint64) {
	infoEventWithPegoutID(pegoutID).Msg("Aggregate sign shares")
}

func (s *SignService) logSignatureSent(pegoutID uint64) {
	infoEventWithPegoutID(pegoutID).Msg("Signature sent")
}

func (s *SignService) logSignatureSendError(err error) {
	errorEvent().Err(err).Msg("failed to send signatures")
}

func (s *SignService) logAggregateSignSharesError(err error) {
	errorEvent().Err(err).Msg("failed to aggregate sign shares")
}
