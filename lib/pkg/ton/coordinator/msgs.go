package coordinator

import (
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func BuildSendRound1Body(ttl int64, validatorIdx uint16, dkgUntil int64, round1Package []byte, r2PublicX25519 *[32]byte) *cell.Cell {
	return cell.BeginCell().
		MustStoreUInt(OpCodeCoordinatorRound1, 32).
		MustStoreUInt(uint64(time.Now().Unix()+ttl), 32).
		MustStoreUInt(uint64(validatorIdx), 16).
		MustStoreUInt(uint64(dkgUntil), 32).
		MustStoreRef(utils.SplitBytesToCells(append(r2PublicX25519[:], round1Package...))).
		EndCell()
}

func BuildSendRound2Body(ttl int64, validatorIdx uint16, dkgUntil int64, round2Packages []byte) *cell.Cell {
	return cell.BeginCell().
		MustStoreUInt(OpCodeCoordinatorRound2, 32).
		MustStoreUInt(uint64(time.Now().Unix()+ttl), 32).
		MustStoreUInt(uint64(validatorIdx), 16).
		MustStoreUInt(uint64(dkgUntil), 32).
		MustStoreRef(utils.SplitBytesToCells(round2Packages)).
		EndCell()
}

func BuildSendRound3Body(ttl int64, validatorIdx uint16, dkgUntil int64, sessionPublicKey []byte, pubkeyPackage []byte) *cell.Cell {
	return cell.BeginCell().
		MustStoreUInt(OpCodeCoordinatorRound3, 32).
		MustStoreUInt(uint64(time.Now().Unix()+ttl), 32).
		MustStoreUInt(uint64(validatorIdx), 16).
		MustStoreUInt(uint64(dkgUntil), 32).
		MustStoreSlice(sessionPublicKey, 256).
		MustStoreRef(utils.SplitBytesToCells(pubkeyPackage)).
		EndCell()
}

func BuildSendDKGClaimBody(ttl int64, validatorIdx uint16, dkgUntil int64, culpritIdx uint16) *cell.Cell {
	return cell.BeginCell().
		MustStoreUInt(OpCodeCoordinatorDkgClaim, 32).
		MustStoreUInt(uint64(time.Now().Unix()+ttl), 32).
		MustStoreUInt(uint64(validatorIdx), 16).
		MustStoreUInt(uint64(dkgUntil), 32).
		MustStoreUInt(uint64(culpritIdx), 16).
		EndCell()
}

func BuildSendCommitmentsBody(ttl int64, req *CommitmentRequest) *cell.Cell {
	return cell.BeginCell().
		MustStoreUInt(OpCodeCoordinatorSendCommitments, 32).
		MustStoreUInt(uint64(time.Now().Unix()+ttl), 32).
		MustStoreUInt(uint64(req.ValidatorIdx), 16).
		MustStoreUInt(req.PegoutID, 64).
		MustStoreRef(utils.SplitBytesToCells(req.Commitments)).
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
		MustStoreUInt(req.PegoutID, 64).
		MustStoreRef(dict.AsCell()).
		EndCell()
}

func BuildSendSignaturesBody(ttl int64, req *SignaturesRequest) *cell.Cell {
	dict := cell.NewDict(16)

	for i, signature := range req.Signatures {
		dict.Set(cell.BeginCell().MustStoreUInt(uint64(i), 16).EndCell(),
			utils.SplitBytesToCells(signature),
		)
	}

	return cell.BeginCell().
		MustStoreUInt(OpCodeCoordinatorSendSignature, 32).
		MustStoreUInt(uint64(time.Now().Unix()+ttl), 32).
		MustStoreUInt(uint64(req.ValidatorIdx), 16).
		MustStoreUInt(req.PegoutID, 64).
		MustStoreRef(dict.AsCell()).
		EndCell()
}

func BuildSendSigningClaimBody(ttl int64, req *SigningClaimRequest) *cell.Cell {
	return cell.BeginCell().
		MustStoreUInt(OpCodeCoordinatorSigningClaim, 32).
		MustStoreUInt(uint64(time.Now().Unix()+ttl), 32).
		MustStoreUInt(uint64(req.ValidatorIdx), 16).
		MustStoreUInt(req.PegoutID, 64).
		MustStoreUInt(uint64(req.culpritIdx), 16).
		EndCell()
}

func BuildSendResetPegoutSigningBody(ttl int64, req *ResetPegoutSigningRequest) *cell.Cell {
	return cell.BeginCell().
		MustStoreUInt(OpCodeCoordinatorResetPegoutSigning, 32).
		MustStoreUInt(uint64(time.Now().Unix()+ttl), 32).
		MustStoreUInt(uint64(req.ValidatorIdx), 16).
		MustStoreUInt(req.PegoutID, 64).
		EndCell()
}
