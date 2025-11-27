package teleportcontract

import (
	"fmt"
	"math/big"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

const (
	eventIdMint   = 0x77a80ef3
	eventIdBurn   = 0xca444ce6
	eventIdReinit = 0x27756729
)

type EventWithPegoutInterface interface {
	ton.EventInterface
	GetAmount() *big.Int
	GetPegoutAddr() *address.Address
}

type EventWithPegout struct {
	Raw        *ton.RawEvent
	Amount     *big.Int
	PegoutAddr *address.Address
}

type MintEvent struct {
	Raw          *ton.RawEvent
	Amount       *big.Int
	ReceiverAddr *address.Address
	BitcoinTxID  *chainhash.Hash
}

type BurnEvent struct {
	EventWithPegout
	SenderAddr *address.Address
}

type ReinitEvent struct {
	EventWithPegout
	NewInternalKey []byte
}

func (e *EventWithPegout) GetAmount() *big.Int             { return e.Amount }
func (e *EventWithPegout) GetPegoutAddr() *address.Address { return e.PegoutAddr }
func (e *MintEvent) GetEventID() uint32                    { return eventIdMint }
func (e *BurnEvent) GetEventID() uint32                    { return eventIdBurn }
func (e *ReinitEvent) GetEventID() uint32                  { return eventIdReinit }
func (e *MintEvent) GetRaw() *ton.RawEvent                 { return e.Raw }
func (e *BurnEvent) GetRaw() *ton.RawEvent                 { return e.Raw }
func (e *ReinitEvent) GetRaw() *ton.RawEvent               { return e.Raw }

type EventParser struct{}

func NewEventParser() *EventParser {
	return &EventParser{}
}

func (ep *EventParser) Parse(raw *ton.RawEvent) (ton.EventInterface, error) {
	s := raw.Body.BeginParse()
	eventId := s.MustLoadUInt(32)

	switch eventId {
	case eventIdMint:
		return ep.parseMintEvent(s, raw)
	case eventIdBurn:
		return ep.parseBurnEvent(s, raw)
	case eventIdReinit:
		return ep.parseReinitEvent(s, raw)
	default:
		return nil, fmt.Errorf("unknown event type with id %x", uint32(eventId))
	}
}

func (ep *EventParser) parseMintEvent(s *cell.Slice, raw *ton.RawEvent) (*MintEvent, error) {
	amount := s.MustLoadBigCoins()
	receiver := s.MustLoadAddr()
	bitcoinTxID, err := chainhash.NewHash(s.MustLoadSlice(256))
	if err != nil {
		return nil, err
	}
	return &MintEvent{
		raw, amount, receiver, bitcoinTxID,
	}, nil
}

func (ep *EventParser) parseBurnEvent(s *cell.Slice, raw *ton.RawEvent) (*BurnEvent, error) {
	amount := s.MustLoadBigCoins()
	senderAddr := s.MustLoadAddr()
	pegoutAddr := s.MustLoadAddr()
	return &BurnEvent{
		EventWithPegout{raw, amount, pegoutAddr}, senderAddr,
	}, nil
}

func (ep *EventParser) parseReinitEvent(s *cell.Slice, raw *ton.RawEvent) (*ReinitEvent, error) {
	amount := s.MustLoadBigCoins()
	newInternalKey := s.MustLoadSlice(256)
	pegoutAddr := s.MustLoadAddr()
	return &ReinitEvent{
		EventWithPegout{raw, amount, pegoutAddr}, newInternalKey,
	}, nil
}
