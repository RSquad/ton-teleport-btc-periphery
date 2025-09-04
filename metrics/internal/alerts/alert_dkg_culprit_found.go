package alerts

import (
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
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

func (alert *AlertDkgCulpritFound) Check(dataSource AlertDataSource) (Severity, Labels, error) {
	emptyLabels := Labels{
		"culprit_id": "",
		"is_evicted": "",
	}

	// Get DKG
	dkg, err := dataSource.DkgDB()
	if err != nil {
		return SEVERITY_UNKNOWN, emptyLabels, err
	}

	// No DKG
	if dkg == nil {
		alert.DkgUntil = time.Unix(0, 0)
		alert.Labels = emptyLabels
		alert.Severity = SEVERITY_OK
		return alert.Severity, alert.Labels, nil
	}

	// First DKG try
	if alert.DkgUntil.Equal(time.Unix(0, 0)) {
		alert.DkgUntil = dkg.Until
		alert.Labels = emptyLabels
		alert.Severity = SEVERITY_OK
		return alert.Severity, alert.Labels, nil
	}

	// Check for restart
	if alert.DkgUntil.Equal(dkg.Until) {
		// No restart
		return alert.Severity, alert.Labels, nil
	}

	// Get DKG info before restart
	dkgBeforeRestart, err := dataSource.DkgBeforeRestartDB(alert.DkgUntil)
	if err != nil {
		return SEVERITY_UNKNOWN, emptyLabels, err
	}
	alert.DkgUntil = dkg.Until

	// Culprit Id
	culpritId := alert.GetCulpritId(dkgBeforeRestart.VSetMask, dkg.VSetMask)

	if culpritId >= 0 {
		alert.Labels["culprit_id"] = strconv.FormatInt(int64(culpritId), 10)
		alert.Labels["is_evicted"] = "YES"
		alert.Severity = SEVERITY_CRITICAL
	} else if len(dkgBeforeRestart.Claims.Counters) > 0 {
		alert.Labels["culprit_id"] = alert.ExtractNotEvicted(&dkgBeforeRestart.Claims.Counters)
		alert.Labels["is_evicted"] = "NO"
		alert.Severity = SEVERITY_CRITICAL
	} else {
		alert.Labels = emptyLabels
		alert.Severity = SEVERITY_OK
	}

	return alert.Severity, alert.Labels, nil
}

func (alert *AlertDkgCulpritFound) GetCulpritId(beforeMask *big.Int, afterMask *big.Int) int {
	var resMask big.Int
	resMask.AndNot(beforeMask, afterMask)

	if resMask.Cmp(big.NewInt(0)) == 0 {
		return -1
	}

	return int(resMask.TrailingZeroBits())
}

func (alert *AlertDkgCulpritFound) ExtractNotEvicted(counters *coordinator.DKGClaimcounters) string {
	keys := make([]int, 0, len(*counters))
	for k := range *counters {
		keys = append(keys, int(k))
	}

	sort.Ints(keys)

	strKeys := make([]string, 0, len(keys))
	for _, k := range keys {
		strKeys = append(strKeys, strconv.Itoa(k))
	}

	return strings.Join(strKeys, ",")
}
