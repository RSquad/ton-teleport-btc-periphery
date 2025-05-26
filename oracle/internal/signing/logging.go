package signing

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	helpers "github.com/rsquad/ton-teleport-btc-periphery/oracle/internal"
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

func (s *SignService) logOracleEvictedFromSigning(pegoutID uint64) {
	err := fmt.Errorf("the Oracle has been evicted from pegout signing: %x", pegoutID)
	errorEvent().Err(err)
}

func (s *SignService) logErrNoOracleCommitments(pegoutID uint64) {
	err := fmt.Errorf("oracle didn't send commitment and cannot participate in signing for pegout %x", pegoutID)
	errorEvent().Err(err)
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

func (s *SignService) logAggregateSignShares() {
	infoEventWithPegoutID(s.cachedPegout.ID).Msg("Aggregate sign shares")
}

func (s *SignService) logSignatureSent(pegoutID uint64) {
	infoEventWithPegoutID(pegoutID).Msg("Signature sent")
}

func (s *SignService) logSignaturesSent(pegoutID uint64, sentSignsCount uint16, totalCount uint16) {
	infoEventWithPegoutID(pegoutID).Msgf("The signature has already been sent. Waiting for other oracles (ready %d of %d)", sentSignsCount, totalCount)
}

func (s *SignService) logSignatureSendError(pegoutID uint64, err error) {
	msg := helpers.HandleTvmError(err)
	errorEventWithPegoutID(pegoutID).Err(err).Msg("failed to send signatures: " + msg)
}

func (s *SignService) logSignError(inputIndex int, err error) {
	errorEvent().Err(err).Msgf("failed to generate signing share for input %d", inputIndex)
}

func (s *SignService) logAggregateSignSharesError(inputIndex int, err error) {
	errorEvent().Err(err).Msgf("failed to aggregate signature for input %d", inputIndex)
}

func (s *SignService) logSendCommitments(pegoutID uint64) {
	infoEventWithPegoutID(pegoutID).Msg("send commitments")
}

func (s *SignService) logSendSigningShare(pegoutID uint64, signShares [][]byte) {
	infoEventWithPegoutID(pegoutID).Msgf("send %d signing shares", len(signShares))
}

func (s *SignService) logSendCommitmentsError(pegoutID uint64, err error) {
	errorEventWithPegoutID(pegoutID).Err(err).Msg("failed to send commitments")
}

func (s *SignService) logSendSigningShareError(pegoutID uint64, err error) {
	msg := helpers.HandleTvmError(err)
	errorEventWithPegoutID(pegoutID).Err(err).Msg("failed to send signing share: " + msg)
}

func (s *SignService) logExecuteClaim(pegoutID uint64) {
	infoEventWithPegoutID(pegoutID).Msg("execute claim")
}

func (s *SignService) logSendClaim(pegoutID uint64, culpritIdx uint16) {
	infoEventWithPegoutID(pegoutID).Msg(fmt.Sprintf("Send claim, culprit validator idx: %d", culpritIdx))
}

func (s *SignService) logSigningClaimSent(pegoutID uint64) {
	infoEventWithPegoutID(pegoutID).Msg("Signing claim sent")
}

func (s *SignService) logSigningClaimSentError(pegoutID uint64, err error) {
	errorEventWithPegoutID(pegoutID).Err(err).Msg("failed to send signing claim")
}

func (s *SignService) logSendResetPegoutSigning(pegoutID uint64) {
	infoEventWithPegoutID(pegoutID).Msg("Send reset pegout signing")
}

func (s *SignService) logResetPegoutSigningSent(pegoutID uint64) {
	infoEventWithPegoutID(pegoutID).Msg("Reset pegout signing sent")
}

func (s *SignService) logResetPegoutSigningSentError(pegoutID uint64, err error) {
	errorEventWithPegoutID(pegoutID).Err(err).Msg("failed to send reset pegout signing")
}
