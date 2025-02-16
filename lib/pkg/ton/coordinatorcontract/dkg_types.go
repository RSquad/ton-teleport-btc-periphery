package coordinatorcontract

import (
	"math/big"
)

type DKG struct {
	status     DKGStatus
	vSet       *VSet
	maxSigners uint64
	r1         *DKGR1
	r2         *DKGR2
	// r3Pkg  R3Pkg
	// cfgHash    []byte
	// attempts   uint64
	// until      uint64
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
