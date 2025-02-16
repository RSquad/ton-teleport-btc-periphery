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
	eventIdReinit = 0x84d432ba
)

type EventWithPegoutInterface interface {
	ton.EventInterface
	GetID() uint32
	GetAmount() *big.Int
	GetBitcoinScript() []byte
}

type EventWithPegout struct {
	Raw           *ton.RawEvent
	ID            uint32
	Amount        *big.Int
	BitcoinScript []byte
}

type MintEvent struct {
	Raw          *ton.RawEvent
	Amount       *big.Int
	ReceiverAddr *address.Address
	BitcoinTxID  *chainhash.Hash
}

type BurnEvent struct {
	EventWithPegout
	BitcoinTxID *chainhash.Hash
	SenderAddr  *address.Address
}

type ReinitEvent struct {
	EventWithPegout
	BitcoinTxID *chainhash.Hash
}

func (e *EventWithPegout) GetID() uint32            { return e.ID }
func (e *EventWithPegout) GetAmount() *big.Int      { return e.Amount }
func (e *EventWithPegout) GetBitcoinScript() []byte { return e.BitcoinScript }
func (e *MintEvent) GetEventID() uint32             { return eventIdMint }
func (e *BurnEvent) GetEventID() uint32             { return eventIdBurn }
func (e *ReinitEvent) GetEventID() uint32           { return eventIdReinit }
func (e *MintEvent) GetRaw() *ton.RawEvent          { return e.Raw }
func (e *BurnEvent) GetRaw() *ton.RawEvent          { return e.Raw }
func (e *ReinitEvent) GetRaw() *ton.RawEvent        { return e.Raw }

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
		return ep.parseBurnLog(s, raw)
	case eventIdReinit:
		return ep.parseReinitLog(s, raw)
	default:
		return nil, fmt.Errorf("unknown event type with id %x", eventId)
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

func (ep *EventParser) parseBurnLog(s *cell.Slice, raw *ton.RawEvent) (*BurnEvent, error) {
	id := uint32(s.MustLoadUInt(32))
	amount := s.MustLoadBigCoins()
	bitcoinTxID, err := chainhash.NewHash(s.MustLoadSlice(256))
	if err != nil {
		return nil, err
	}
	senderAddr := s.MustLoadAddr()
	bitcoinScriptSlice := s.MustLoadRef()
	bitcoinScript := bitcoinScriptSlice.MustLoadSlice(uint(bitcoinScriptSlice.MustLoadUInt(8) * 8))
	return &BurnEvent{
		EventWithPegout{raw, id, amount, bitcoinScript}, bitcoinTxID, senderAddr,
	}, nil
}

func (ep *EventParser) parseReinitLog(s *cell.Slice, raw *ton.RawEvent) (*ReinitEvent, error) {
	id := uint32(s.MustLoadUInt(32))
	amount := s.MustLoadBigCoins()
	bitcoinTxID, err := chainhash.NewHash(s.MustLoadSlice(256))
	if err != nil {
		return nil, err
	}
	bitcoinScriptSlice := s.MustLoadRef()
	var bitcoinScript []byte
	if bitcoinScriptSlice.BitsLeft() > 2 {
		bitcoinScript = bitcoinScriptSlice.MustLoadSlice(bitcoinScriptSlice.BitsLeft())
	}
	return &ReinitEvent{
		EventWithPegout{raw, id, amount, bitcoinScript}, bitcoinTxID,
	}, nil
}
