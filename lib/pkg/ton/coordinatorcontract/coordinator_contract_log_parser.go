package coordinatorcontract

import (
	"fmt"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

const (
	eventIdDKGComplete = 0x453443a6
)

type DKGCompletedEvent struct {
	Raw         *ton.RawEvent
	CompletedAt time.Time
	Key         []byte
}

func (m *DKGCompletedEvent) GetEventID() uint32 {
	return eventIdDKGComplete
}

func (m *DKGCompletedEvent) GetRaw() *ton.RawEvent {
	return m.Raw
}

type EventParser struct{}

func NewEventParser() *EventParser {
	return &EventParser{}
}

func (ep *EventParser) Parse(raw *ton.RawEvent) (ton.EventInterface, error) {
	s := raw.Body.BeginParse()
	eventId := s.MustLoadUInt(32)

	switch eventId {
	case eventIdDKGComplete:
		return parseDKGCompleteEvent(s, raw)
	default:
		return nil, fmt.Errorf("unknown event type with id %x", eventId)
	}
}

func parseDKGCompleteEvent(s *cell.Slice, raw *ton.RawEvent) (*DKGCompletedEvent, error) {
	completedAt := s.MustLoadBigUInt(64)
	key := s.MustLoadSlice(256)
	return &DKGCompletedEvent{
		raw, time.Unix(completedAt.Int64(), 0), key,
	}, nil
}
