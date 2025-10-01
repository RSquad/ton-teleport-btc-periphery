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
	DkgUntil   time.Time
	Labels     Labels
	Severity   Severity
	CulpritIds map[int]struct{}
}

func NewAlertDkgCulpritFound() Alert {
	return &AlertDkgCulpritFound{
		DkgUntil:   time.Unix(0, 0),
		Labels:     Labels{},
		Severity:   SEVERITY_UNKNOWN,
		CulpritIds: make(map[int]struct{}, 0),
	}
}

func (alert *AlertDkgCulpritFound) NewLabels() Labels {
	return Labels{
		"culprit_ids":     "",
		"not_evicted_ids": "",
	}
}

func (alert *AlertDkgCulpritFound) Check(dataSource AlertDataSource) (Severity, Labels, Values, error) {
	labels := alert.NewLabels()

	// Get DKG
	dkg, err := dataSource.DkgDB()
	if err != nil {
		return SEVERITY_UNKNOWN, labels, alert.MakeValues(), err
	}

	// No DKG
	if dkg == nil {
		alert.DkgUntil = time.Unix(0, 0)
		alert.Labels = labels
		alert.Severity = SEVERITY_OK
		alert.CulpritIds = make(map[int]struct{}, 0)
		return alert.Severity, alert.Labels, alert.MakeValues(), nil
	}

	// First DKG try
	if alert.DkgUntil.Equal(time.Unix(0, 0)) {
		alert.DkgUntil = dkg.Until
		alert.Labels = labels
		alert.Severity = SEVERITY_OK
		alert.CulpritIds = make(map[int]struct{}, 0)
		return alert.Severity, alert.Labels, alert.MakeValues(), nil
	}

	// Check for restart
	if alert.DkgUntil.Equal(dkg.Until) {
		// No restart
		return alert.Severity, alert.Labels, alert.MakeValues(), nil
	}

	// Get DKG info before restart
	dkgBeforeRestart, err := dataSource.DkgBeforeRestartDB(alert.DkgUntil)
	if err != nil {
		return SEVERITY_UNKNOWN, labels, alert.MakeValues(), err
	}
	alert.DkgUntil = dkg.Until

	// Culprit Id or list of not evicted
	if dkgBeforeRestart.Claims != nil && len(dkgBeforeRestart.Claims.Counters) > 0 {
		coordinatorContractData, err := dataSource.CoordinatorContractStorageDB()
		if err != nil {
			return SEVERITY_UNKNOWN, labels, alert.MakeValues(), err
		}

		prevDkg, err := dataSource.PrevDkgDB()
		if err != nil {
			return SEVERITY_UNKNOWN, labels, alert.MakeValues(), err
		}

		culpritId, listOfNotEvicted := alert.Extract(
			&dkgBeforeRestart.Claims.Counters,
			int(coordinatorContractData.MinClaimsPercent),
			int(prevDkg.MaxSigners),
		)

		if culpritId >= 0 {
			alert.CulpritIds[culpritId] = struct{}{}
		}

		culpritIdsStr := mutils.ExtractMapKeysConv(alert.CulpritIds, strconv.Itoa)

		alert.Labels["culprit_ids"] = strings.Join(culpritIdsStr, ",")
		alert.Labels["not_evicted_ids"] = strings.Join(listOfNotEvicted, ",")
		alert.Severity = SEVERITY_CRITICAL
	} else {
		alert.Labels = labels
		alert.Severity = SEVERITY_OK
	}

	return alert.Severity, alert.Labels, alert.MakeValues(), nil
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

func (alert *AlertDkgCulpritFound) MakeValues() Values {
	values := make(Values, 1)

	culpritIds := mutils.ExtractMapKeys(alert.CulpritIds)
	values["culprit_ids"] = culpritIds

	return values
}
