package coordinator

import (
	"context"
	"fmt"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/signer"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func (c *coordinatorContract) SendStartDKG() (*tlb.Transaction, error) {
	unsignedMsgBody := cell.BeginCell().
		MustStoreUInt(OpCodeStartDKG, 32).
		MustStoreUInt(uint64(time.Now().Unix()+int64(c.ttl.Seconds())), 32).
		EndCell()
	msg, err := ton.BuildExtMsg(unsignedMsgBody, c.Addr, c.signer)
	if err != nil {
		return nil, err
	}

	tx, err := CallApiWithTimeout(
		func(apiCallCtx context.Context) (*tlb.Transaction, error) {
			tx, _, _, err := c.tonClient.API.SendExternalMessageWaitTransaction(apiCallCtx, msg)
			return tx, err
		},
		c.ctx,
		c.tonApiCallTimeout,
		"SendExternalMessageWaitTransaction: SendStartDKG",
	)

	return tx, err
}

func (c *coordinatorContract) SendRound1(
	validatorIdx uint16,
	dkgUntil int64,
	round1Package []byte,
	r2PublicX25519 *[32]byte,
) (*tlb.Transaction, error) {
	return c.sendBodyCell(BuildSendRound1Body(
		int64(c.ttl.Seconds()), validatorIdx, dkgUntil, round1Package, r2PublicX25519,
	), "SendRound1")
}

func (c *coordinatorContract) SendRound2(
	validatorIdx uint16,
	dkgUntil int64,
	round2Packages []byte,
) (*tlb.Transaction, error) {
	return c.sendBodyCell(BuildSendRound2Body(
		int64(c.ttl.Seconds()), validatorIdx, dkgUntil, round2Packages,
	), "SendRound2")
}

func (c *coordinatorContract) SendDKGClaim(
	validatorIdx uint16,
	dkgUntil int64,
	culpritIdx uint16,
) (*tlb.Transaction, error) {
	return c.sendBodyCell(BuildSendDKGClaimBody(
		int64(c.ttl.Seconds()), validatorIdx, dkgUntil, culpritIdx,
	), "SendDKGClaim")
}

func (c *coordinatorContract) SendPubkeyPackage(
	validatorIdx uint16,
	dkgUntil int64,
	sessionPublicKey []byte,
	pubkeyPackage []byte,
) (*tlb.Transaction, error) {
	return c.sendBodyCell(BuildSendRound3Body(
		int64(c.ttl.Seconds()), validatorIdx, dkgUntil, sessionPublicKey, pubkeyPackage,
	), "SendPubkeyPackage")
}

func (c *coordinatorContract) SendCommitments(
	PegoutID uint64,
	PegoutUntil int64,
	ValidatorIdx uint16,
	Commitments []byte,
) (*tlb.Transaction, error) {
	if len(Commitments) == 0 {
		return nil, fmt.Errorf("commitments are empty")
	}
	return c.sendBodyCell(BuildSendCommitmentsBody(
		int64(c.ttl.Seconds()),
		&CommitmentRequest{PegoutID, PegoutUntil, ValidatorIdx, Commitments},
	), "SendCommitments")
}

func (c *coordinatorContract) SendSigningShare(
	PegoutID uint64,
	PegoutUntil int64,
	ValidatorIdx uint16,
	SigningShares [][]byte,
) (*tlb.Transaction, error) {
	return c.sendBodyCell(BuildSendSigningShareBody(
		int64(c.ttl.Seconds()),
		&SigningShareRequest{PegoutID, PegoutUntil, ValidatorIdx, SigningShares},
	), "SendSigningShare")
}

func (c *coordinatorContract) SendSignatures(
	PegoutID uint64,
	PegoutUntil int64,
	ValidatorIdx uint16,
	Signatures [][]byte,
) (*tlb.Transaction, error) {
	return c.sendBodyCell(BuildSendSignaturesBody(
		int64(c.ttl.Seconds()),
		&SignaturesRequest{PegoutID, PegoutUntil, ValidatorIdx, Signatures},
	), "SendSignatures")
}

func (c *coordinatorContract) SendSigningClaim(
	PegoutID uint64,
	PegoutUntil int64,
	ValidatorIdx uint16,
	culpritIdx uint16,
) (*tlb.Transaction, error) {
	return c.sendBodyCell(BuildSendSigningClaimBody(
		int64(c.ttl.Seconds()),
		&SigningClaimRequest{PegoutID, PegoutUntil, ValidatorIdx, culpritIdx},
	), "SendSigningClaim")
}

func (c *coordinatorContract) SendResetPegoutSigning(
	PegoutID uint64,
	ValidatorIdx uint16,
) (*tlb.Transaction, error) {
	return c.sendBodyCell(BuildSendResetPegoutSigningBody(
		int64(c.ttl.Seconds()),
		&ResetPegoutSigningRequest{PegoutID, ValidatorIdx},
	), "SendResetPegoutSigning")
}

func (c *coordinatorContract) ConnectSigner(signer signer.Signer) {
	c.signer = signer
}

func (c *coordinatorContract) sendBodyCell(bodyCell *cell.Cell, name string) (*tlb.Transaction, error) {
	msg, err := ton.BuildExtMsg(bodyCell, c.Addr, c.signer)
	if err != nil {
		return nil, err
	}

	tx, err := CallApiWithTimeout(
		func(apiCallCtx context.Context) (*tlb.Transaction, error) {
			tx, _, _, err := c.tonClient.API.SendExternalMessageWaitTransaction(apiCallCtx, msg)
			return tx, err
		},
		c.ctx,
		c.tonApiCallTimeout,
		name,
	)

	return tx, err
}
