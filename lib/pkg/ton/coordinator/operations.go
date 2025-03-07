package coordinator

import (
	"fmt"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/signer"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func (c *CoordinatorContract) SendStartDKG(ttl int64) (*tlb.Transaction, error) {
	if ttl == 0 {
		ttl = int64(DefaultDGKTTL.Seconds())
	}

	unsignedMsgBody := cell.BeginCell().
		MustStoreUInt(OpCodeStartDKG, 32).
		MustStoreUInt(uint64(time.Now().Unix()+ttl), 32).
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

func (c *CoordinatorContract) SendPubkeyPackage(validatorIdx uint16, internalKeyXY []byte, Identifier []byte, pubkeyPackage []byte) (*tlb.Transaction, error) {
	if (len(internalKeyXY) != 65) || (internalKeyXY[0] != 0x04) {
		return nil, fmt.Errorf("internalKeyXY must be 65 bytes and has prefix 0x04")
	}
	return c.sendBodyCell(BuildSendRound3Body(
		int64(c.ttl.Seconds()), validatorIdx, internalKeyXY, Identifier, pubkeyPackage,
	))
}

func (c *CoordinatorContract) SendCommitments(req *CommitmentRequest) (*tlb.Transaction, error) {
	return c.sendBodyCell(BuildSendCommitmentsBody(int64(c.ttl.Seconds()), req))
}

func (c *CoordinatorContract) SendSigningShare(req *SigningShareRequest) (*tlb.Transaction, error) {
	return c.sendBodyCell(BuildSendSigningShareBody(int64(c.ttl.Seconds()), req))
}

func (c *CoordinatorContract) SendSignatures(req *SignaturesRequest) (*tlb.Transaction, error) {
	return c.sendBodyCell(BuildSendSignaturesBody(int64(c.ttl.Seconds()), req))
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
