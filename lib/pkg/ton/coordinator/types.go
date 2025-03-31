package coordinator

import (
	"encoding/hex"
	"math/big"
	"time"

	"github.com/xssnick/tonutils-go/address"
)

type DKG struct {
	Status     DKGStatus
	VSet       VSet
	VsetMask   []byte
	MaxSigners uint16
	R1         *DKGR1
	R2         *DKGR2
	R3         *DKGR3
	Claims     *DKGClaims
	CfgHash    []byte
	Attempts   uint64
	Until      time.Time
}

type DKGRoundState struct {
	mask  *big.Int
	count uint64
}

type DKGStatus uint64

const (
	DKGStatusFinished      DKGStatus = 0
	DKGStatusInProgress    DKGStatus = 1
	DKGStatusPart1Finished DKGStatus = 2
	DKGStatusPart2Finished DKGStatus = 3
)

func (s DKGStatus) String() string {
	switch s {
	case DKGStatusFinished:
		return "FINISHED"
	case DKGStatusInProgress:
		return "IN_PROGRESS"
	case DKGStatusPart1Finished:
		return "PART1_FINISHED"
	case DKGStatusPart2Finished:
		return "PART2_FINISHED"
	default:
		return "UNKNOWN"
	}
}

func (dkg *DKG) GetR1Packages() DKGPkgs {
	return dkg.R1.Packages
}

func (dkg *DKG) GetR2Packages(fromIdentifier []byte) DKGPkgs {
	return dkg.R2.Packages[hex.EncodeToString(fromIdentifier)]
}

func (dkg *DKG) Round1Completed() bool {
	return dkg.Status == DKGStatusFinished ||
		dkg.Status >= DKGStatusPart1Finished
}

func (dkg *DKG) Round2Completed() bool {
	return dkg.Status == DKGStatusFinished ||
		dkg.Status >= DKGStatusPart2Finished
}

func (dkg *DKG) Round3Completed() bool {
	return dkg.Status == DKGStatusFinished
}

func (dkg *DKG) ClaimCompleted(validatorIdx uint16) bool {
	return dkg.Claims.Mask.Bit(int(validatorIdx)) > 0
}

// CommitmentRequest represents a request to send commitments
type CommitmentRequest struct {
	PegoutID     uint64
	ValidatorIdx uint16
	Identifier   []byte
	Commitments  []byte
}

type SigningShareRequest struct {
	PegoutID      uint64
	ValidatorIdx  uint16
	Identifier    []byte
	SigningShares [][]byte
}

type SignaturesRequest struct {
	PegoutID     uint64
	ValidatorIdx uint16
	Signatures   [][]byte
}

type PegoutRecord struct {
	ID                uint64
	PegoutAddress     *address.Address
	InternalKey       []byte
	Commitments       map[string][]byte
	CommitmentsMask   []byte
	SigningShares     map[string]map[int][]byte
	SigningSharesMask []byte
}

func (p *PegoutRecord) HasCommitment(identifier []byte) bool {
	_, exists := p.Commitments[hex.EncodeToString(identifier)]
	return exists
}

func (p *PegoutRecord) CommitmentsCount() uint16 {
	return uint16(len(p.Commitments))
}

func (p *PegoutRecord) HasSigningShare(identifier []byte) bool {
	_, exists := p.SigningShares[hex.EncodeToString(identifier)]
	return exists
}

func (p *PegoutRecord) SigningSharesCount() int {
	return len(p.SigningShares)
}
