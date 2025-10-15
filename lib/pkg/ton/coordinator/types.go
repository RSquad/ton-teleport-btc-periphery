package coordinator

import (
	"math/big"
	"time"

	"github.com/xssnick/tonutils-go/address"
)

type DKG struct {
	State       DKGState
	VSet        VSet
	MaxSigners  uint16
	VSetMask    *big.Int
	SessionKeys *SessionKeys
	R1          *DKGR1
	R2          *DKGR2
	R3          *DKGR3
	Claims      *DKGClaims
	CfgHash     []byte
	Attempts    uint64
	Until       time.Time
}

type DKGRoundState struct {
	mask  *big.Int
	count uint64
}

type DKGState uint64

const (
	DKGStateFinished      DKGState = 0
	DKGStateInProgress    DKGState = 1
	DKGStatePart1Finished DKGState = 2
	DKGStatePart2Finished DKGState = 3
)

func (s DKGState) String() string {
	switch s {
	case DKGStateFinished:
		return "FINISHED"
	case DKGStateInProgress:
		return "IN_PROGRESS"
	case DKGStatePart1Finished:
		return "PART1_FINISHED"
	case DKGStatePart2Finished:
		return "PART2_FINISHED"
	default:
		return "UNKNOWN"
	}
}

func (dkg *DKG) GetR1Packages() DKGPkgs {
	return dkg.R1.Packages
}

func (dkg *DKG) GetR2Packages() DKGPkgs {
	return dkg.R2.Packages
}

func (dkg *DKG) Round1Completed() bool {
	return dkg.State == DKGStateFinished ||
		dkg.State >= DKGStatePart1Finished
}

func (dkg *DKG) Round2Completed() bool {
	return dkg.State == DKGStateFinished ||
		dkg.State >= DKGStatePart2Finished
}

func (dkg *DKG) Round3Completed() bool {
	return dkg.State == DKGStateFinished
}

func (dkg *DKG) CheckR1Mask(validatorIdx uint16) (bool, uint16) {
	return dkg.R1.Mask.Bit(int(validatorIdx)) > 0, uint16(dkg.R1.Count)
}

func (dkg *DKG) CheckR2Mask(validatorIdx uint16) (bool, uint16) {
	return dkg.R2.Mask.Bit(int(validatorIdx)) > 0, uint16(dkg.R2.Count)
}

func (dkg *DKG) CheckR3Mask(validatorIdx uint16) (bool, uint16) {
	return dkg.R3.Mask.Bit(int(validatorIdx)) > 0, uint16(dkg.R3.Count)
}

func (dkg *DKG) ClaimCompleted(validatorIdx uint16) bool {
	return dkg.Claims.Mask.Bit(int(validatorIdx)) > 0
}

func (dkg *DKG) CheckVSetMask(validatorIdx uint16) bool {
	return dkg.VSetMask.Bit(int(validatorIdx)) > 0
}

// CommitmentRequest represents a request to send commitments
type CommitmentRequest struct {
	PegoutID     uint64
	PegoutUntil  int64
	ValidatorIdx uint16
	Commitments  []byte
}

type SigningShareRequest struct {
	PegoutID      uint64
	PegoutUntil   int64
	ValidatorIdx  uint16
	SigningShares [][]byte
}

type SignaturesRequest struct {
	PegoutID     uint64
	PegoutUntil  int64
	ValidatorIdx uint16
	Signatures   [][]byte
}

type SigningClaimRequest struct {
	PegoutID     uint64
	PegoutUntil  int64
	ValidatorIdx uint16
	culpritIdx   uint16
}

type ResetPegoutSigningRequest struct {
	PegoutID     uint64
	ValidatorIdx uint16
}

type PegoutSignatures struct {
	Mask  *big.Int
	Count uint16
	Hash  []byte
}

type PegoutRecord struct {
	ID                      uint64
	PegoutAddress           *address.Address // address for the pegout contract
	InternalKey             []byte           // internal publickey used to sign pegout inputs
	IsAutopegout            bool             // true if pegout is autopegout
	Commitments             map[uint16][]byte
	CommitmentsMaskAccepted *big.Int
	CommitmentsMaskOther    *big.Int
	SigningShares           map[uint16]map[uint16][]byte
	SigningSharesMask       []byte
	Signatures              PegoutSignatures
	ClaimsMask              *big.Int
	ClaimsCount             uint16
	ClaimsCounters          map[uint16]uint16
	Signers                 uint16    // number of signers (count of bit 1 in signing mask)
	ExpiredAt               time.Time // timestamp when the pegout expires
	SigningMask             *big.Int  // bitmask for signing permissions
}

func (p *PegoutRecord) HasCommitment(idx uint16) bool {
	_, exists := p.Commitments[idx]
	return exists
}

func (p *PegoutRecord) HasCommitmentOther(idx uint16) bool {
	return p.CommitmentsMaskOther.Bit(int(idx)) != 0
}

func (p *PegoutRecord) CommitmentsCount() uint16 {
	return uint16(len(p.Commitments))
}

func (p *PegoutRecord) HasSigningShare(idx uint16) bool {
	_, exists := p.SigningShares[idx]
	return exists
}

func (p *PegoutRecord) SigningSharesCount() int {
	return len(p.SigningShares)
}

func (p *PegoutRecord) CheckSigningMask(validatorIdx uint16) bool {
	return p.SigningMask.Bit(int(validatorIdx)) > 0
}
