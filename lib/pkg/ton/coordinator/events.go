package coordinator

import (
	"fmt"
	"math/big"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type EventId uint

const (
	eventIdDKGComplete            EventId = 0x453443a6
	eventIdDKGStarted             EventId = 0x1ab2bd7f
	eventIdDKGCompletedInfo       EventId = 0x4e062aa5
	eventIdDKGRestarted           EventId = 0xf661074e
	eventIdDKGRotated             EventId = 0x3f78a410
	eventIdPegoutSigningStarted   EventId = 0xa19457a2
	eventIdPegoutSigningCompleted EventId = 0x317d48b4
	eventIdPegoutSigningRestarted EventId = 0x82280c5c
)

type DKGCompletedEvent struct {
	Raw         *ton.RawEvent
	CompletedAt time.Time
	Key         []byte
}

type DKGStartedEvent struct {
	Raw *ton.RawEvent
}

type DKGCompletedInfoEvent struct {
	Raw *ton.RawEvent
}

type DKGRestartedEvent struct {
	Reason     *big.Int
	NewDkg     *cell.Slice
	Claims     DKGClaimcounters
	ClaimsMask *big.Int
	Raw        *ton.RawEvent
}

type DKGRotatedEvent struct {
	Raw *ton.RawEvent
}

type PegoutSigningStartedEvent struct {
	PegoutId *big.Int
	Raw      *ton.RawEvent
}

type PegoutSigningCompletedEvent struct {
	PegoutId *big.Int
	Raw      *ton.RawEvent
}

type PegoutSigningRestartedEvent struct {
	PegoutId       *big.Int
	Reason         *big.Int
	Pegout         *cell.Slice
	CommitmentMask *big.Int
	SharesMask     *big.Int
	SignatureMask  *big.Int
	Claims         DKGClaimcounters
	ClaimsMask     *big.Int
	Raw            *ton.RawEvent
}

func (m *DKGCompletedEvent) GetEventID() uint32 {
	return uint32(eventIdDKGComplete)
}

func (m *DKGStartedEvent) GetEventID() uint32 {
	return uint32(eventIdDKGStarted)
}

func (m *DKGCompletedInfoEvent) GetEventID() uint32 {
	return uint32(eventIdDKGCompletedInfo)
}

func (m *DKGRestartedEvent) GetEventID() uint32 {
	return uint32(eventIdDKGRestarted)
}

func (m *DKGRotatedEvent) GetEventID() uint32 {
	return uint32(eventIdDKGRotated)
}

func (m *PegoutSigningStartedEvent) GetEventID() uint32 {
	return uint32(eventIdPegoutSigningStarted)
}

func (m *PegoutSigningCompletedEvent) GetEventID() uint32 {
	return uint32(eventIdPegoutSigningCompleted)
}

func (m *PegoutSigningRestartedEvent) GetEventID() uint32 {
	return uint32(eventIdPegoutSigningRestarted)
}

func (m *DKGCompletedEvent) GetRaw() *ton.RawEvent {
	return m.Raw
}

func (m *DKGStartedEvent) GetRaw() *ton.RawEvent {
	return m.Raw
}

func (m *DKGCompletedInfoEvent) GetRaw() *ton.RawEvent {
	return m.Raw
}

func (m *DKGRestartedEvent) GetRaw() *ton.RawEvent {
	return m.Raw
}

func (m *DKGRotatedEvent) GetRaw() *ton.RawEvent {
	return m.Raw
}

func (m *PegoutSigningStartedEvent) GetRaw() *ton.RawEvent {
	return m.Raw
}

func (m *PegoutSigningCompletedEvent) GetRaw() *ton.RawEvent {
	return m.Raw
}

func (m *PegoutSigningRestartedEvent) GetRaw() *ton.RawEvent {
	return m.Raw
}

type EventParser struct{}

func NewEventParser() *EventParser {
	return &EventParser{}
}

func (ep *EventParser) Parse(raw *ton.RawEvent) (ton.EventInterface, error) {
	s := raw.Body.BeginParse()
	eventId := EventId(s.MustLoadUInt(32))

	switch eventId {
	case eventIdDKGComplete:
		return parseDKGCompleteEvent(s, raw)
	case eventIdDKGStarted:
		return parseDKGStartedEvent(raw)
	case eventIdDKGCompletedInfo:
		return parseDKGCompletedInfoEvent(raw)
	case eventIdDKGRestarted:
		return parseDKGRestartedEvent(s, raw)
	case eventIdDKGRotated:
		return parseDKGRotatedEvent(raw)
	case eventIdPegoutSigningStarted:
		return parsePegoutSigningStartedEvent(s, raw)
	case eventIdPegoutSigningCompleted:
		return parsePegoutSigningCompletedEvent(s, raw)
	case eventIdPegoutSigningRestarted:
		return parsePegoutSigningRestartedEvent(s, raw)
	default:
		return nil, fmt.Errorf("unknown event type with id %x", uint32(eventId))
	}
}

func parseDKGCompleteEvent(s *cell.Slice, raw *ton.RawEvent) (*DKGCompletedEvent, error) {
	completedAt := s.MustLoadBigUInt(64)
	key := s.MustLoadSlice(256)
	return &DKGCompletedEvent{
		raw, time.Unix(completedAt.Int64(), 0), key,
	}, nil
}

func parseDKGStartedEvent(raw *ton.RawEvent) (*DKGStartedEvent, error) {
	return &DKGStartedEvent{
		raw,
	}, nil
}

func parseDKGCompletedInfoEvent(raw *ton.RawEvent) (*DKGCompletedInfoEvent, error) {
	return &DKGCompletedInfoEvent{
		raw,
	}, nil
}

func parseDKGRestartedEvent(s *cell.Slice, raw *ton.RawEvent) (*DKGRestartedEvent, error) {
	reason := s.MustLoadBigUInt(8)
	newDkg := s.MustLoadRef()
	claims, err := NewDKGClaimcounters(s.MustLoadDict(16))
	if err != nil {
		return nil, err
	}
	claimsMask := s.MustLoadBigUInt(256)

	return &DKGRestartedEvent{
		reason, newDkg, claims, claimsMask,
		raw,
	}, nil
}

func parseDKGRotatedEvent(raw *ton.RawEvent) (*DKGRotatedEvent, error) {
	return &DKGRotatedEvent{
		raw,
	}, nil
}

func parsePegoutSigningStartedEvent(s *cell.Slice, raw *ton.RawEvent) (*PegoutSigningStartedEvent, error) {
	pegoutId := s.MustLoadBigUInt(64)
	return &PegoutSigningStartedEvent{
		pegoutId,
		raw,
	}, nil
}

func parsePegoutSigningCompletedEvent(s *cell.Slice, raw *ton.RawEvent) (*PegoutSigningCompletedEvent, error) {
	pegoutId := s.MustLoadBigUInt(64)
	return &PegoutSigningCompletedEvent{
		pegoutId,
		raw,
	}, nil
}

func parsePegoutSigningRestartedEvent(s *cell.Slice, raw *ton.RawEvent) (*PegoutSigningRestartedEvent, error) {
	pegoutId := s.MustLoadBigUInt(64)
	reason := s.MustLoadBigUInt(8)
	pegout := s.MustLoadRef()
	claims, err := NewDKGClaimcounters(s.MustLoadDict(16))
	claimsMask := s.MustLoadBigUInt(256)
	s = s.MustLoadRef()
	commitmentMask := s.MustLoadBigUInt(256)
	sharesMask := s.MustLoadBigUInt(256)
	signatureMask := s.MustLoadBigUInt(256)
	if err != nil {
		return nil, err
	}

	return &PegoutSigningRestartedEvent{
		pegoutId, reason, pegout, commitmentMask, sharesMask, signatureMask, claims, claimsMask,
		raw,
	}, nil
}
