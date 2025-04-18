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

func (dkg *DKG) CheckR1Mask(validatorIdx uint16) bool {
	return dkg.R1.mask.Bit(int(validatorIdx)) > 0
}

func (dkg *DKG) CheckR2Mask(validatorIdx uint16) bool {
	return dkg.R2.mask.Bit(int(validatorIdx)) > 0
}

func (dkg *DKG) CheckR3Mask(validatorIdx uint16) bool {
	return dkg.R1.mask.Bit(int(validatorIdx)) > 0
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
	ValidatorIdx uint16
	Commitments  []byte
}

type SigningShareRequest struct {
	PegoutID      uint64
	ValidatorIdx  uint16
	SigningShares [][]byte
}

type SignaturesRequest struct {
	PegoutID     uint64
	ValidatorIdx uint16
	Signatures   [][]byte
}

type SigningClaimRequest struct {
	PegoutID     uint64
	ValidatorIdx uint16
	culpritIdx   uint16
}

type ResetPegoutSigningRequest struct {
	PegoutID     uint64
	ValidatorIdx uint16
}

type PegoutSignatures struct {
	mask  *big.Int
	count uint16
	hash  []byte
}

type PegoutRecord struct {
	ID                uint64
	PegoutAddress     *address.Address
	InternalKey       []byte
	Commitments       map[uint16][]byte
	CommitmentsMask   []byte
	SigningShares     map[uint16]map[uint16][]byte
	SigningSharesMask []byte
	Signatures        PegoutSignatures
	ClaimsMask        *big.Int
	ClaimsCount       uint16
	ClaimsCounters    map[uint16]uint16
	MaxSigners        uint16
	ExpiredAt         time.Time
	SigningMask       *big.Int
}

func (p *PegoutRecord) HasCommitment(idx uint16) bool {
	_, exists := p.Commitments[idx]
	return exists
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
