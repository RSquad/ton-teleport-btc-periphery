package coordinator

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

const (
	EventIdDKGComplete            = 0x453443a6 // 1161053094
	EventIdDKGStarted             = 0x1ab2bd7f // 447921535
	EventIdDKGCompletedInfo       = 0x4e062aa5 // 1309026981
	EventIdDKGRestarted           = 0xf661074e // 4133553998
	EventIdDKGRotated             = 0x3f78a410 // 1064870928
	EventIdPegoutSigningStarted   = 0xa19457a2 // 2710853538
	EventIdPegoutSigningCompleted = 0x317d48b4 // 830294196
	EventIdPegoutSigningRestarted = 0x82280c5c // 2183662684
)

type DkgRestartReason = int

const (
	DkgRestartTimeoutExpired   DkgRestartReason = 1
	DkgRestartValidatorEvicted DkgRestartReason = 2
)

type DKGCompletedEvent struct {
	Raw         *ton.RawEvent `json:"-"`
	CompletedAt time.Time
	Key         []byte
}

type DKGStartedEvent struct {
	Raw *ton.RawEvent `json:"-"`
	Dkg *DKG
}

type DKGCompletedInfoEvent struct {
	Raw *ton.RawEvent `json:"-"`
	Dkg *DKG
}

type DKGRestartedEvent struct {
	Raw        *ton.RawEvent `json:"-"`
	Reason     DkgRestartReason
	NewDkg     *DKG
	Claims     DKGClaimcounters
	ClaimsMask *big.Int
}

type DKGRotatedEvent struct {
	Raw *ton.RawEvent `json:"-"`
}

type PegoutSigningStartedEvent struct {
	Raw      *ton.RawEvent `json:"-"`
	PegoutId uint64
	Pegout   *PegoutRecord
}

type PegoutSigningCompletedEvent struct {
	Raw      *ton.RawEvent `json:"-"`
	PegoutId uint64
	Pegout   *PegoutRecord
}

type PegoutSigningRestartedEvent struct {
	Raw            *ton.RawEvent `json:"-"`
	PegoutId       uint64
	Reason         uint8
	Pegout         *PegoutRecord
	CommitmentMask *big.Int
	SharesMask     *big.Int
	SignatureMask  *big.Int
	Claims         DKGClaimcounters
	ClaimsMask     *big.Int
}

func (m *DKGCompletedEvent) GetEventID() uint32 {
	return EventIdDKGComplete
}

func (m *DKGStartedEvent) GetEventID() uint32 {
	return EventIdDKGStarted
}

func (m *DKGCompletedInfoEvent) GetEventID() uint32 {
	return EventIdDKGCompletedInfo
}

func (m *DKGRestartedEvent) GetEventID() uint32 {
	return EventIdDKGRestarted
}

func (m *DKGRotatedEvent) GetEventID() uint32 {
	return EventIdDKGRotated
}

func (m *PegoutSigningStartedEvent) GetEventID() uint32 {
	return EventIdPegoutSigningStarted
}

func (m *PegoutSigningCompletedEvent) GetEventID() uint32 {
	return EventIdPegoutSigningCompleted
}

func (m *PegoutSigningRestartedEvent) GetEventID() uint32 {
	return EventIdPegoutSigningRestarted
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

func (m *DKGCompletedEvent) SetRaw(raw *ton.RawEvent) {
	m.Raw = raw
}

func (m *DKGStartedEvent) SetRaw(raw *ton.RawEvent) {
	m.Raw = raw
}

func (m *DKGCompletedInfoEvent) SetRaw(raw *ton.RawEvent) {
	m.Raw = raw
}

func (m *DKGRestartedEvent) SetRaw(raw *ton.RawEvent) {
	m.Raw = raw
}

func (m *DKGRotatedEvent) SetRaw(raw *ton.RawEvent) {
	m.Raw = raw
}

func (m *PegoutSigningStartedEvent) SetRaw(raw *ton.RawEvent) {
	m.Raw = raw
}

func (m *PegoutSigningCompletedEvent) SetRaw(raw *ton.RawEvent) {
	m.Raw = raw
}

func (m *PegoutSigningRestartedEvent) SetRaw(raw *ton.RawEvent) {
	m.Raw = raw
}

type EventParser struct{}

func NewEventParser() *EventParser {
	return &EventParser{}
}

func (ep *EventParser) Parse(raw *ton.RawEvent) (ton.EventInterface, error) {
	s := raw.Body.BeginParse()
	eventId, err := s.LoadUInt(32)
	if err != nil {
		return nil, err
	}

	switch eventId {
	case EventIdDKGComplete:
		return ParseDKGCompleteEvent(s, raw)
	case EventIdDKGStarted:
		return ParseDKGStartedEvent(s, raw)
	case EventIdDKGCompletedInfo:
		return ParseDKGCompletedInfoEvent(s, raw)
	case EventIdDKGRestarted:
		return ParseDKGRestartedEvent(s, raw)
	case EventIdDKGRotated:
		return ParseDKGRotatedEvent(s, raw)
	case EventIdPegoutSigningStarted:
		return ParsePegoutSigningStartedEvent(s, raw)
	case EventIdPegoutSigningCompleted:
		return ParsePegoutSigningCompletedEvent(s, raw)
	case EventIdPegoutSigningRestarted:
		return ParsePegoutSigningRestartedEvent(s, raw)
	default:
		return nil, fmt.Errorf("unknown event type with id %x", eventId)
	}
}

func (ep *EventParser) ParseJson(jsonStr []byte, eventId uint64) (ton.EventInterface, error) {
	switch eventId {
	case EventIdDKGComplete:
		var ev DKGCompletedEvent
		err := json.Unmarshal(jsonStr, &ev)
		return &ev, err
	case EventIdDKGStarted:
		var ev DKGStartedEvent
		err := json.Unmarshal(jsonStr, &ev)
		return &ev, err
	case EventIdDKGCompletedInfo:
		var ev DKGCompletedInfoEvent
		err := json.Unmarshal(jsonStr, &ev)
		return &ev, err
	case EventIdDKGRestarted:
		var ev DKGRestartedEvent
		err := json.Unmarshal(jsonStr, &ev)
		return &ev, err
	case EventIdDKGRotated:
		var ev DKGRotatedEvent
		err := json.Unmarshal(jsonStr, &ev)
		return &ev, err
	case EventIdPegoutSigningStarted:
		var ev PegoutSigningStartedEvent
		err := json.Unmarshal(jsonStr, &ev)
		return &ev, err
	case EventIdPegoutSigningCompleted:
		var ev PegoutSigningCompletedEvent
		err := json.Unmarshal(jsonStr, &ev)
		return &ev, err
	case EventIdPegoutSigningRestarted:
		var ev PegoutSigningRestartedEvent
		err := json.Unmarshal(jsonStr, &ev)
		return &ev, err
	default:
		return nil, fmt.Errorf("unknown event type with id %x", eventId)
	}
}

func ParseDKGCompleteEvent(s *cell.Slice, raw *ton.RawEvent) (*DKGCompletedEvent, error) {
	completedAt, err := s.LoadBigUInt(64)
	if err != nil {
		return nil, err
	}

	key, err := s.LoadSlice(256)
	if err != nil {
		return nil, err
	}

	return &DKGCompletedEvent{
		Raw:         raw,
		CompletedAt: time.Unix(completedAt.Int64(), 0),
		Key:         key,
	}, nil
}

func ParseDKGStartedEvent(s *cell.Slice, raw *ton.RawEvent) (*DKGStartedEvent, error) {
	dkgSlice, err := s.LoadRef()
	if err != nil {
		return nil, err
	}

	dkg, err := parseDGKSlice(dkgSlice)
	if err != nil {
		return nil, err
	}

	return &DKGStartedEvent{
		Raw: raw,
		Dkg: dkg,
	}, nil
}

func ParseDKGCompletedInfoEvent(s *cell.Slice, raw *ton.RawEvent) (*DKGCompletedInfoEvent, error) {
	dkgSlice, err := s.LoadRef()
	if err != nil {
		return nil, err
	}

	dkg, err := parseDGKSlice(dkgSlice)
	if err != nil {
		return nil, err
	}

	return &DKGCompletedInfoEvent{
		Raw: raw,
		Dkg: dkg,
	}, nil
}

func ParseDKGRestartedEvent(s *cell.Slice, raw *ton.RawEvent) (*DKGRestartedEvent, error) {
	reasonBig, err := s.LoadBigUInt(8)
	if err != nil {
		return nil, err
	}

	reason, err := utils.BigToUint8(reasonBig)
	if err != nil {
		return nil, err
	}

	newDkgSlice, err := s.LoadRef()
	if err != nil {
		return nil, err
	}

	newDkg, err := parseDGKSlice(newDkgSlice)
	if err != nil {
		return nil, err
	}

	claimsDict, err := s.LoadDict(16)
	if err != nil {
		return nil, err
	}

	claims, err := NewDKGClaimcounters(claimsDict)
	if err != nil {
		return nil, err
	}

	claimsMask, err := s.LoadBigUInt(256)
	if err != nil {
		return nil, err
	}

	return &DKGRestartedEvent{
		Raw:        raw,
		Reason:     DkgRestartReason(reason),
		NewDkg:     newDkg,
		Claims:     claims,
		ClaimsMask: claimsMask,
	}, nil
}

func ParseDKGRotatedEvent(s *cell.Slice, raw *ton.RawEvent) (*DKGRotatedEvent, error) {
	return &DKGRotatedEvent{
		raw,
	}, nil
}

func ParsePegoutSigningStartedEvent(s *cell.Slice, raw *ton.RawEvent) (*PegoutSigningStartedEvent, error) {
	pegoutIdBig, err := s.LoadBigUInt(64)
	if err != nil {
		return nil, err
	}

	pegoutId, err := utils.BigToUint64(pegoutIdBig)
	if err != nil {
		return nil, err
	}

	pegoutSlice, err := s.LoadRef()
	if err != nil {
		return nil, err
	}

	pegout, err := parsePegout(pegoutSlice, pegoutId)
	if err != nil {
		return nil, err
	}

	return &PegoutSigningStartedEvent{
		Raw:      raw,
		PegoutId: pegoutId,
		Pegout:   pegout,
	}, nil
}

func ParsePegoutSigningCompletedEvent(s *cell.Slice, raw *ton.RawEvent) (*PegoutSigningCompletedEvent, error) {
	pegoutIdBig, err := s.LoadBigUInt(64)
	if err != nil {
		return nil, err
	}

	pegoutId, err := utils.BigToUint64(pegoutIdBig)
	if err != nil {
		return nil, err
	}

	pegoutSlice, err := s.LoadRef()
	if err != nil {
		return nil, err
	}

	pegout, err := parsePegout(pegoutSlice, pegoutId)
	if err != nil {
		return nil, err
	}

	return &PegoutSigningCompletedEvent{
		Raw:      raw,
		PegoutId: pegoutId,
		Pegout:   pegout,
	}, nil
}

func ParsePegoutSigningRestartedEvent(s *cell.Slice, raw *ton.RawEvent) (*PegoutSigningRestartedEvent, error) {
	pegoutIdBig, err := s.LoadBigUInt(64)
	if err != nil {
		return nil, err
	}

	pegoutId, err := utils.BigToUint64(pegoutIdBig)
	if err != nil {
		return nil, err
	}

	reasonBig, err := s.LoadBigUInt(8)
	if err != nil {
		return nil, err
	}

	reason, err := utils.BigToUint8(reasonBig)
	if err != nil {
		return nil, err
	}

	pegoutSlice, err := s.LoadRef()
	if err != nil {
		return nil, err
	}

	pegout, err := parsePegout(pegoutSlice, pegoutId)
	if err != nil {
		return nil, err
	}

	other, err := s.LoadRef()
	if err != nil {
		return nil, err
	}

	claimsDict, err := s.LoadDict(16)
	if err != nil {
		return nil, err
	}

	claims, err := NewDKGClaimcounters(claimsDict)
	if err != nil {
		return nil, err
	}

	claimsMask, err := s.LoadBigUInt(256)
	if err != nil {
		return nil, err
	}

	commitmentMask, err := other.LoadBigUInt(256)
	if err != nil {
		return nil, err
	}

	sharesMask, err := other.LoadBigUInt(256)
	if err != nil {
		return nil, err
	}

	signatureMask, err := other.LoadBigUInt(256)
	if err != nil {
		return nil, err
	}

	return &PegoutSigningRestartedEvent{
		Raw:            raw,
		PegoutId:       pegoutId,
		Reason:         reason,
		Pegout:         pegout,
		CommitmentMask: commitmentMask,
		SharesMask:     sharesMask,
		SignatureMask:  signatureMask,
		Claims:         claims,
		ClaimsMask:     claimsMask,
	}, nil
}
