package coordinatorcontract

import (
	"math/big"
	"time"
)

type DKG struct {
	Status     DKGStatus
	VSet       *VSet
	MaxSigners uint64
	R1         *DKGR1
	R2         *DKGR2
	Until      time.Time
	// r3Pkg  R3Pkg
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
