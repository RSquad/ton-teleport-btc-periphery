package coordinatorcontract

import (
	"fmt"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

const DefaultDGKTTL = time.Minute * 5

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

func (c *CoordinatorContract) SendRound1(ttl int64, validatorIdx uint16, Identifier []byte, round1Package []byte) (*tlb.Transaction, error) {
	if ttl == 0 {
		ttl = int64(DefaultDGKTTL.Seconds())
	}

	unsignedMsgBody := cell.BeginCell().
		MustStoreUInt(OpCodeCoordinatorRound1, 32).
		MustStoreUInt(uint64(time.Now().Unix()+ttl), 32).
		MustStoreUInt(uint64(validatorIdx), 16).
		MustStoreRef(
			cell.BeginCell().
				MustStoreSlice(Identifier, 32).
				MustStoreRef(
					utils.SplitBytesToCells(round1Package),
				).
				EndCell(),
		).
		EndCell()

	msg, err := ton.BuildExtMsg(unsignedMsgBody, c.Addr, c.signer)
	if err != nil {
		return nil, err
	}

	tx, _, _, err := c.tonClient.API.SendExternalMessageWaitTransaction(c.ctx, msg)

	return tx, err
}

func (c *CoordinatorContract) SendRound2(ttl int64, validatorIdx uint16, fromIdentifier []byte, toIdentifier []byte, round2Package []byte) (*tlb.Transaction, error) {
	if ttl == 0 {
		ttl = int64(DefaultDGKTTL.Seconds())
	}

	unsignedMsgBody := cell.BeginCell().
		MustStoreUInt(OpCodeCoordinatorRound2, 32).
		MustStoreUInt(uint64(time.Now().Unix()+ttl), 32).
		MustStoreUInt(uint64(validatorIdx), 16).
		MustStoreRef(
			cell.BeginCell().
				MustStoreSlice(fromIdentifier, 32).
				MustStoreSlice(toIdentifier, 32).
				MustStoreRef(
					utils.SplitBytesToCells(round2Package),
				).
				EndCell(),
		).
		EndCell()

	msg, err := ton.BuildExtMsg(unsignedMsgBody, c.Addr, c.signer)
	if err != nil {
		return nil, err
	}

	tx, _, _, err := c.tonClient.API.SendExternalMessageWaitTransaction(c.ctx, msg)

	return tx, err
}

func (c *CoordinatorContract) SendPubkeyPackage(ttl int64, validatorIdx uint16, internalKeyXY []byte, Identifier []byte, pubkeyPackage []byte) (*tlb.Transaction, error) {
	if ttl == 0 {
		ttl = int64(DefaultDGKTTL.Seconds())
	}

	if (len(internalKeyXY) != 65) || (internalKeyXY[0] != 0x04) {
		return nil, fmt.Errorf("internalKeyXY must be 65 bytes and has prefix 0x04")
	}

	unsignedMsgBody := cell.BeginCell().
		MustStoreUInt(OpCodeCoordinatorRound3, 32).
		MustStoreUInt(uint64(time.Now().Unix()+ttl), 32).
		MustStoreUInt(uint64(validatorIdx), 16).
		MustStoreRef(
			cell.BeginCell().
				MustStoreSlice(Identifier, 32).
				MustStoreSlice(internalKeyXY[1:65], 64).
				MustStoreRef(
					utils.SplitBytesToCells(pubkeyPackage),
				).
				EndCell(),
		).
		EndCell()

	msg, err := ton.BuildExtMsg(unsignedMsgBody, c.Addr, c.signer)
	if err != nil {
		return nil, err
	}

	tx, _, _, err := c.tonClient.API.SendExternalMessageWaitTransaction(c.ctx, msg)

	return tx, err
}
