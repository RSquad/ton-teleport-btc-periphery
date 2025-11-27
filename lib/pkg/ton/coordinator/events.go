package coordinator

import (
	"fmt"
	"math/big"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

const (
	eventIdDKGComplete            = 0x453443a6
	eventIdDKGStarted             = 0x1ab2bd7f
	eventIdDKGCompletedInfo       = 0x4e062aa5
	eventIdDKGRestarted           = 0xf661074e
	eventIdDKGRotated             = 0x3f78a410
	eventIdPegoutSigningStarted   = 0xa19457a2
	eventIdPegoutSigningCompleted = 0x317d48b4
	eventIdPegoutSigningRestarted = 0x82280c5c
)

type DKGCompletedEvent struct {
	Raw         *ton.RawEvent
	CompletedAt time.Time
	Key         []byte
}

type DKGStartedEvent struct {
	Raw *ton.RawEvent
	Dkg *cell.Slice
}

type DKGCompletedInfoEvent struct {
	Raw *ton.RawEvent
	Dkg *cell.Slice
}

type DKGRestartedEvent struct {
	Raw        *ton.RawEvent
	Reason     uint8
	NewDkg     *cell.Slice
	Claims     DKGClaimcounters
	ClaimsMask *big.Int
}

type DKGRotatedEvent struct {
	Raw *ton.RawEvent
}

type PegoutSigningStartedEvent struct {
	Raw      *ton.RawEvent
	PegoutId *big.Int
	Pegout   *cell.Slice
}

type PegoutSigningCompletedEvent struct {
	Raw      *ton.RawEvent
	PegoutId *big.Int
	Pegout   *cell.Slice
}

type PegoutSigningRestartedEvent struct {
	Raw            *ton.RawEvent
	PegoutId       uint64
	Reason         uint8
	Pegout         *cell.Slice
	CommitmentMask *big.Int
	SharesMask     *big.Int
	SignatureMask  *big.Int
	Claims         DKGClaimcounters
	ClaimsMask     *big.Int
}

func (m *DKGCompletedEvent) GetEventID() uint32 {
	return eventIdDKGComplete
}

func (m *DKGStartedEvent) GetEventID() uint32 {
	return eventIdDKGStarted
}

func (m *DKGCompletedInfoEvent) GetEventID() uint32 {
	return eventIdDKGCompletedInfo
}

func (m *DKGRestartedEvent) GetEventID() uint32 {
	return eventIdDKGRestarted
}

func (m *DKGRotatedEvent) GetEventID() uint32 {
	return eventIdDKGRotated
}

func (m *PegoutSigningStartedEvent) GetEventID() uint32 {
	return eventIdPegoutSigningStarted
}

func (m *PegoutSigningCompletedEvent) GetEventID() uint32 {
	return eventIdPegoutSigningCompleted
}

func (m *PegoutSigningRestartedEvent) GetEventID() uint32 {
	return eventIdPegoutSigningRestarted
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
	eventId := s.MustLoadUInt(32)

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
