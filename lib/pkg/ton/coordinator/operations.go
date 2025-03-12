package coordinator

import (
	"fmt"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/signer"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func (c *CoordinatorContract) SendStartDKG() (*tlb.Transaction, error) {
	unsignedMsgBody := cell.BeginCell().
		MustStoreUInt(OpCodeStartDKG, 32).
		MustStoreUInt(uint64(time.Now().Unix()+int64(c.ttl.Seconds())), 32).
		EndCell()
	msg, err := ton.BuildExtMsg(unsignedMsgBody, c.Addr, c.signer)
	if err != nil {
		return nil, err
	}

	tx, _, _, err := c.tonClient.API.SendExternalMessageWaitTransaction(c.ctx, msg)

	return tx, err
}

func (c *CoordinatorContract) SendRound1(validatorIdx uint16, identifier []byte, round1Package []byte) (*tlb.Transaction, error) {
	return c.sendBodyCell(BuildSendRound1Body(
		int64(c.ttl.Seconds()), validatorIdx, identifier, round1Package,
	))
}

func (c *CoordinatorContract) SendRound2(validatorIdx uint16, fromIdentifier []byte, toIdentifier []byte, round2Package []byte) (*tlb.Transaction, error) {
	return c.sendBodyCell(BuildSendRound2Body(
		int64(c.ttl.Seconds()), validatorIdx, fromIdentifier, toIdentifier, round2Package,
	))
}

func (c *CoordinatorContract) SendPubkeyPackage(validatorIdx uint16, internalKeyX []byte, Identifier []byte, pubkeyPackage []byte) (*tlb.Transaction, error) {
	if len(internalKeyX) != 32 {
		return nil, fmt.Errorf("internalKeyX must be 32 bytes long, got %d", len(internalKeyX))
	}
	return c.sendBodyCell(BuildSendRound3Body(
		int64(c.ttl.Seconds()), validatorIdx, internalKeyX, Identifier, pubkeyPackage,
	))
}

func (c *CoordinatorContract) SendCommitments(
	PegoutID uint64,
	ValidatorIdx uint16,
	Identifier, Commitments []byte,
) (*tlb.Transaction, error) {
	return c.sendBodyCell(BuildSendCommitmentsBody(
		int64(c.ttl.Seconds()),
		&CommitmentRequest{PegoutID, ValidatorIdx, Identifier, Commitments},
	))
}

func (c *CoordinatorContract) SendSigningShare(
	PegoutID uint64,
	ValidatorIdx uint16,
	Identifier []byte,
	SigningShares [][]byte,
) (*tlb.Transaction, error) {
	return c.sendBodyCell(BuildSendSigningShareBody(
		int64(c.ttl.Seconds()),
		&SigningShareRequest{PegoutID, ValidatorIdx, Identifier, SigningShares},
	))
}

func (c *CoordinatorContract) SendSignatures(
	PegoutID uint64,
	ValidatorIdx uint16,
	Identifier []byte,
	Signatures [][]byte,
) (*tlb.Transaction, error) {
	return c.sendBodyCell(BuildSendSignaturesBody(
		int64(c.ttl.Seconds()),
		&SignaturesRequest{PegoutID, ValidatorIdx, Identifier, Signatures},
	))
}

func (c *CoordinatorContract) ConnectSigner(signer *signer.Signer) {
	c.signer = signer
}

func (c *CoordinatorContract) sendBodyCell(bodyCell *cell.Cell) (*tlb.Transaction, error) {
	msg, err := ton.BuildExtMsg(bodyCell, c.Addr, c.signer)
	if err != nil {
		return nil, err
	}
	tx, _, _, err := c.tonClient.API.SendExternalMessageWaitTransaction(c.ctx, msg)
	return tx, err
}
