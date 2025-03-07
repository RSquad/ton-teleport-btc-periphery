package coordinatorcontract

import (
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func BuildSendRound1Body(ttl int64, validatorIdx uint16, Identifier []byte, round1Package []byte) *cell.Cell {
	return cell.BeginCell().
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
}

func BuildSendRound2Body(ttl int64, validatorIdx uint16, fromIdentifier []byte, toIdentifier []byte, round2Package []byte) *cell.Cell {
	return cell.BeginCell().
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
}

func BuildSendRound3Body(ttl int64, validatorIdx uint16, internalKeyXY []byte, Identifier []byte, pubkeyPackage []byte) *cell.Cell {
	return cell.BeginCell().
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
}
