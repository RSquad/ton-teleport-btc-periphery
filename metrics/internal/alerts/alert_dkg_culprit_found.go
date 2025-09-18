package alerts

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
)

type AlertDkgCulpritFound struct {
	DkgUntil time.Time
	Labels   Labels
	Severity Severity
}

func NewAlertDkgCulpritFound() Alert {
	return &AlertDkgCulpritFound{
		DkgUntil: time.Unix(0, 0),
		Labels:   Labels{},
		Severity: SEVERITY_UNKNOWN,
	}
}

func (alert *AlertDkgCulpritFound) NewLabels() Labels {
	return Labels{
		"culprit_id":      "",
		"not_evicted_ids": "",
	}
}

func (alert *AlertDkgCulpritFound) Check(dataSource AlertDataSource) (Severity, Labels, IntValues, error) {
	labels := alert.NewLabels()

	// Get DKG
	dkg, err := dataSource.DkgDB()
	if err != nil {
		return SEVERITY_UNKNOWN, labels, nil, err
	}

	// No DKG
	if dkg == nil {
		alert.DkgUntil = time.Unix(0, 0)
		alert.Labels = labels
		alert.Severity = SEVERITY_OK
		return alert.Severity, alert.Labels, nil, nil
	}

	// First DKG try
	if alert.DkgUntil.Equal(time.Unix(0, 0)) {
		alert.DkgUntil = dkg.Until
		alert.Labels = labels
		alert.Severity = SEVERITY_OK
		return alert.Severity, alert.Labels, nil, nil
	}

	// Check for restart
	if alert.DkgUntil.Equal(dkg.Until) {
		// No restart
		return alert.Severity, alert.Labels, nil, nil
	}

	// Get DKG info before restart
	dkgBeforeRestart, err := dataSource.DkgBeforeRestartDB(alert.DkgUntil)
	if err != nil {
		return SEVERITY_UNKNOWN, labels, nil, err
	}
	alert.DkgUntil = dkg.Until

	// Culprit Id or list of not evicted
	if dkgBeforeRestart.Claims != nil && len(dkgBeforeRestart.Claims.Counters) > 0 {
		coordinatorContractData, err := dataSource.CoordinatorContractStorageDB()
		if err != nil {
			return SEVERITY_UNKNOWN, labels, nil, err
		}

		prevDkg, err := dataSource.PrevDkgDB()
		if err != nil {
			return SEVERITY_UNKNOWN, labels, nil, err
		}

		culpritId, listOfNotEvicted := alert.Extract(
			&dkgBeforeRestart.Claims.Counters,
			int(coordinatorContractData.MinClaimsPercent),
			int(prevDkg.MaxSigners),
		)

		if culpritId >= 0 {
			alert.Labels["culprit_id"] = strconv.FormatInt(int64(culpritId), 10)
		} else {
			alert.Labels["culprit_id"] = ""
		}
		alert.Labels["not_evicted_ids"] = strings.Join(listOfNotEvicted, ",")
		alert.Severity = SEVERITY_CRITICAL
	} else {
		alert.Labels = labels
		alert.Severity = SEVERITY_OK
	}

	return alert.Severity, alert.Labels, nil, nil
}

func (alert *AlertDkgCulpritFound) Extract(
	counters *coordinator.DKGClaimcounters,
	minClaimsPercent int,
	maxSigners int,
) (int, []string) {
	culpritId := -1
	listOfNotEvicted := make([]string, 0)

	for idx, votesCount := range *counters {
		percentage := mutils.MulDivCeil(uint(votesCount), 100, uint(maxSigners))
		if int(percentage) >= minClaimsPercent {
			culpritId = int(idx)
		} else {
			listOfNotEvicted = append(listOfNotEvicted, strconv.Itoa(int(idx)))
		}
	}

	if len(listOfNotEvicted) > 0 {
		sort.Strings(listOfNotEvicted)
	}

	return culpritId, listOfNotEvicted
}
