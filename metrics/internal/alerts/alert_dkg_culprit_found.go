package alerts

import (
	"math/big"
	"strconv"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
)

type AlertDkgCulpritFound struct {
	CurrentMask *big.Int
	DkgState    coordinator.DKGState
}

func NewAlertDkgCulpritFound() Alert {
	return &AlertDkgCulpritFound{
		CurrentMask: big.NewInt(0),
		DkgState:    coordinator.DKGStateFinished,
	}
}

func (alert *AlertDkgCulpritFound) Check(dataSource AlertDataSource) (Severity, Labels, error) {
	labels := Labels{
		"culprit_id": "",
		"is_evicted": "",
	}

	// Get DKG
	dkg, err := dataSource.DkgDB()
	if err != nil {
		return SEVERITY_UNKNOWN, labels, err
	}

	if dkg == nil || dkg.Claims == nil {
		alert.CurrentMask = big.NewInt(0)
		alert.DkgState = coordinator.DKGStateFinished
		return SEVERITY_OK, labels, nil
	}

	// New DKG state (reset claim stat)
	if alert.DkgState != dkg.State {
		alert.DkgState = dkg.State
		alert.CurrentMask = big.NewInt(0)
	}

	claimsCount := len(dkg.Claims.Counters)
	evictedCount := 0

	if claimsCount == 0 {
		return SEVERITY_OK, labels, nil
	}

	maxSigners := dkg.MaxSigners

	// Get coordinatior contract storage
	coordinatorStorage, err := dataSource.CoordinatorContractStorageDB()
	if err != nil {
		return SEVERITY_UNKNOWN, labels, err
	}
	minClaimsPercent := uint(coordinatorStorage.MinClaimsPercent)
	claimsMask := dkg.Claims.Mask

	for idx, votesCount := range dkg.Claims.Counters {
		claimsMask = claimsMask.SetBit(claimsMask, int(idx), 1)
		votesPercent := mutils.MulDivCeil(uint(votesCount), 100, uint(maxSigners))
		if votesPercent >= minClaimsPercent {
			evictedCount++
		}
	}

	// Update labels
	var mask big.Int
	mask.AndNot(claimsMask, alert.CurrentMask)
	labels["culprit_id"] = strconv.FormatUint(uint64(mask.TrailingZeroBits()), 10)

	if claimsCount != evictedCount {
		labels["is_evicted"] = "NO"
	} else {
		labels["is_evicted"] = "YES"
	}

	alert.CurrentMask = dkg.Claims.Mask

	return SEVERITY_CRITICAL, labels, nil
}
