package coordinator

import (
	"math/big"
	"time"

	"github.com/xssnick/tonutils-go/address"
)

type DKG struct {
	Status     DKGStatus
	VSet       VSet
	MaxSigners uint64
	R1         *DKGR1
	R2         *DKGR2
	Until      time.Time
	R3         *DKGR3
	// cfgHash    []byte
	// attempts   uint64
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
	return dkg.R2.Packages[string(fromIdentifier)]
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
	Identifier   []byte
	Signatures   [][]byte
}

type PegoutRecord struct {
	ID                uint64
	PegoutAddress     *address.Address
	InternalKey       []byte
	Commitments       map[string][]byte
	CommitmentsMask   []byte
	SigningShares     map[string]map[uint8][]byte
	SigningSharesMask []byte
}

func (p *PegoutRecord) HasCommitment(identifier []byte) bool {
	_, exists := p.Commitments[string(identifier)]
	return exists
}

func (p *PegoutRecord) CommitmentsCount() int {
	return len(p.Commitments)
}

func (p *PegoutRecord) HasSigningShare(identifier []byte) bool {
	_, exists := p.SigningShares[string(identifier)]
	return exists
}

func (p *PegoutRecord) SigningSharesCount() int {
	return len(p.SigningShares)
}
