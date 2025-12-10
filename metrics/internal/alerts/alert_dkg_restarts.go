package alerts

import (
	"errors"
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
)

type AlertDkgRestarts struct{}

type info struct {
	restartsCount                int
	culpritsCount                int
	timeoutEvictedCount          int
	evictedIds                   []uint16
	allEvictedIds                []uint16
	allEvictedAddr               map[uint16][]byte
	fromSecondRestartEvictedAddr map[uint16][]byte
	participantsCount            int
}

func newInfo() *info {
	return &info{
		restartsCount:                0,
		culpritsCount:                0,
		timeoutEvictedCount:          0,
		evictedIds:                   nil,
		allEvictedIds:                nil,
		allEvictedAddr:               nil,
		fromSecondRestartEvictedAddr: nil,
		participantsCount:            0,
	}
}

func NewAlertDkgRestarts() Alert {
	return &AlertDkgRestarts{}
}

func (alert *AlertDkgRestarts) Check(dataSource AlertDataSource) (Severity, Description, Values, error) {
	// Get last DKG Start event
	eventDkgStarted, err := dataSource.EventsLastDkgStartedDB()
	if err != nil {
		return SEVERITY_CRITICAL, "", alert.makeValues(newInfo()), err
	}

	// Chec if no start events
	if eventDkgStarted == nil {
		return SEVERITY_OK, "OK", alert.makeValues(newInfo()), nil
	}

	// Get all DKG Restart events after DKG Start event
	eventDkgRestarts, err := dataSource.EventsAllFromDkgRestartDB(eventDkgStarted.GetRaw().TxLT)
	if err != nil {
		return SEVERITY_CRITICAL, "", alert.makeValues(newInfo()), err
	}

	// Update alert status
	info, err := alert.getInfo(eventDkgRestarts, eventDkgStarted)
	if err != nil {
		return SEVERITY_CRITICAL, "", alert.makeValues(newInfo()), err
	}

	severity := alert.getSeverity(info)
	description := alert.getDescription(severity, info)
	values := alert.makeValues(info)

	return severity, description, values, nil
}

func (alert *AlertDkgRestarts) getInfo(eventDkgRestarts []*coordinator.DKGRestartedEvent, eventDkgStart *coordinator.DKGStartedEvent) (*info, error) {
	restartsCount := len(eventDkgRestarts)
	culpritsCount := 0
	timeoutEvictedCount := 0

	startVSetMask := eventDkgStart.Dkg.VSetMask
	prevVSetMask := startVSetMask
	nowVSetMask := startVSetMask
	prevVSetMaskPopcnt := mutils.Popcnt(prevVSetMask)
	nowVSetMaskPopcnt := prevVSetMaskPopcnt

	for _, restartEvent := range eventDkgRestarts {
		if restartEvent == nil {
			return nil, errors.New("restartEvent is nil")
		}

		if restartEvent.NewDkg == nil {
			return nil, errors.New("restartEvent.NewDkg is nil")
		}

		if restartEvent.NewDkg.VSetMask == nil {
			return nil, errors.New("restartEvent.NewDkg.VSetMask is nil")
		}

		nowVSetMask := restartEvent.NewDkg.VSetMask
		nowVSetMaskPopcnt = mutils.Popcnt(nowVSetMask)
		evictedCount := prevVSetMaskPopcnt - nowVSetMaskPopcnt

		if restartEvent.Reason == coordinator.DkgRestartTimeoutExpired {
			timeoutEvictedCount += evictedCount
		} else if restartEvent.Reason == coordinator.DkgRestartValidatorEvicted {
			culpritsCount += evictedCount
		} else {
			return nil, fmt.Errorf("unknown DKG restart reason %d", restartEvent.Reason)
		}

		prevVSetMask = nowVSetMask
		prevVSetMaskPopcnt = nowVSetMaskPopcnt
	}

	evictedIds := mutils.RemovedOneIds(prevVSetMask, nowVSetMask)
	allEvictedIds := mutils.RemovedOneIds(startVSetMask, nowVSetMask)
	allEvictedAddr := mutils.ExtractValuesByIdx(allEvictedIds, eventDkgStart.Dkg.VSet)

	var fromSecondRestartEvictedAddr map[uint16][]byte = nil
	if restartsCount > 1 {
		fromSecondRestartEvictedIds := mutils.RemovedOneIds(eventDkgRestarts[1].NewDkg.VSetMask, nowVSetMask)
		fromSecondRestartEvictedAddr = mutils.ExtractValuesByIdx(fromSecondRestartEvictedIds, eventDkgStart.Dkg.VSet)
	}

	return &info{
		restartsCount:                restartsCount,
		culpritsCount:                culpritsCount,
		timeoutEvictedCount:          timeoutEvictedCount,
		evictedIds:                   evictedIds,
		allEvictedIds:                allEvictedIds,
		allEvictedAddr:               allEvictedAddr,
		fromSecondRestartEvictedAddr: fromSecondRestartEvictedAddr,
		participantsCount:            nowVSetMaskPopcnt,
	}, nil
}

func (alert *AlertDkgRestarts) getSeverity(inf *info) Severity {
	severity := SEVERITY_OK

	// TODO: implement inf.participantsCount check

	if (inf.restartsCount >= 5) || (inf.culpritsCount > 0) {
		severity = SEVERITY_CRITICAL
	} else if inf.restartsCount >= 2 {
		severity = SEVERITY_WARNING
	}

	return severity
}

func (alert *AlertDkgRestarts) getDescription(severity Severity, inf *info) Description {
	description := "OK"

	if severity > SEVERITY_OK {
		description = fmt.Sprintf(
			"The DKG was restarted %d times.\n<b>Runbook url:</b> %s",
			inf.restartsCount,
			mutils.RunbookLink("DKGRestarts"),
		)
	}

	return Description(description)
}

func (alert *AlertDkgRestarts) makeValues(inf *info) Values {
	values := make(Values, 3)
	values["restartsCount"] = int64(inf.restartsCount)
	values["culpritsCount"] = int64(inf.culpritsCount)
	values["timeoutEvictedCount"] = int64(inf.timeoutEvictedCount)
	values["evictedIds"] = inf.evictedIds
	values["allEvictedIds"] = inf.allEvictedIds
	values["allEvictedAddr"] = inf.allEvictedAddr
	values["fromSecondRestartEvictedAddr"] = inf.fromSecondRestartEvictedAddr
	values["participantsCount"] = inf.participantsCount

	return values
}
