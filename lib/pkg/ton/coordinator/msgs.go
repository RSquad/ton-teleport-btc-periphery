package coordinator

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
				MustStoreSlice(Identifier, 256).
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
				MustStoreSlice(fromIdentifier, 256).
				MustStoreSlice(toIdentifier, 256).
				MustStoreRef(
					utils.SplitBytesToCells(round2Package),
				).
				EndCell(),
		).
		EndCell()
}

func BuildSendRound3Body(ttl int64, validatorIdx uint16, internalKeyX []byte, Identifier []byte, pubkeyPackage []byte) *cell.Cell {
	return cell.BeginCell().
		MustStoreUInt(OpCodeCoordinatorRound3, 32).
		MustStoreUInt(uint64(time.Now().Unix()+ttl), 32).
		MustStoreUInt(uint64(validatorIdx), 16).
		MustStoreRef(
			cell.BeginCell().
				MustStoreSlice(Identifier, 256).
				MustStoreSlice(internalKeyX, 256).
				MustStoreRef(
					utils.SplitBytesToCells(pubkeyPackage),
				).
				EndCell(),
		).
		EndCell()
}

func BuildSendCommitmentsBody(ttl int64, req *CommitmentRequest) *cell.Cell {
	return cell.BeginCell().
		MustStoreUInt(OpCodeCoordinatorSendCommitments, 32).
		MustStoreUInt(uint64(time.Now().Unix()+ttl), 32).
		MustStoreUInt(uint64(req.ValidatorIdx), 16).
		MustStoreRef(
			cell.BeginCell().
				MustStoreSlice(req.Identifier, 32).
				MustStoreUInt(req.PegoutID, 64).
				MustStoreRef(
					utils.SplitBytesToCells(req.Commitments),
				).
				EndCell(),
		).
		EndCell()
}

func BuildSendSigningShareBody(ttl int64, req *SigningShareRequest) *cell.Cell {
	dict := cell.NewDict(64)

	for i, share := range req.SigningShares {
		dict.Set(cell.BeginCell().MustStoreUInt(uint64(i), 64).EndCell(),
			cell.BeginCell().MustStoreRef(utils.SplitBytesToCells(share)).EndCell(),
		)
	}
	return cell.BeginCell().
		MustStoreUInt(OpCodeCoordinatorSendSigningShare, 32).
		MustStoreUInt(uint64(time.Now().Unix()+ttl), 32).
		MustStoreUInt(uint64(req.ValidatorIdx), 16).
		MustStoreRef(
			cell.BeginCell().
				MustStoreSlice(req.Identifier, 32).
				MustStoreUInt(req.PegoutID, 64).
				MustStoreDict(dict).
				EndCell(),
		).
		EndCell()
}

func BuildSendSignaturesBody(ttl int64, req *SignaturesRequest) *cell.Cell {
	dict := cell.NewDict(16)

	for i, signature := range req.Signatures {
		dict.Set(cell.BeginCell().MustStoreUInt(uint64(i), 64).EndCell(),
			utils.SplitBytesToCells(signature),
		)
	}

	return cell.BeginCell().
		MustStoreUInt(OpCodeCoordinatorSendSignature, 32).
		MustStoreUInt(uint64(time.Now().Unix()+ttl), 32).
		MustStoreUInt(uint64(req.ValidatorIdx), 16).
		MustStoreRef(
			cell.BeginCell().
				MustStoreSlice(req.Identifier, 32).
				MustStoreUInt(req.PegoutID, 32).
				MustStoreDict(dict).
				EndCell(),
		).
		EndCell()
}
