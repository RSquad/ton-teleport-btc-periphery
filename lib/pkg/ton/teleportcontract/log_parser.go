package teleportcontract

import (
	"fmt"
	"math/big"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

const (
	logIdMint   = 0x77a80ef3
	logIdBurn   = 0xca444ce6
	logIdReinit = 0x84d432ba
)

type LogInterface interface {
	GetLogID() uint32
}

type LogWithPegoutInterface interface {
	LogInterface
	GetID() uint32
	GetAmount() *big.Int
	GetBitcoinScript() []byte
}

type LogWithPegout struct {
	ID            uint32
	Amount        *big.Int
	BitcoinScript []byte
}

type MintLog struct {
	Amount       *big.Int
	ReceiverAddr *address.Address
	BitcoinTxID  *chainhash.Hash
}

type BurnLog struct {
	LogWithPegout
	BitcoinTxID *chainhash.Hash
	SenderAddr  *address.Address
}

type ReinitLog struct {
	LogWithPegout
	BitcoinTxID *chainhash.Hash
}

func (l *LogWithPegout) GetID() uint32            { return l.ID }
func (l *LogWithPegout) GetAmount() *big.Int      { return l.Amount }
func (l *LogWithPegout) GetBitcoinScript() []byte { return l.BitcoinScript }
func (l *MintLog) GetLogID() uint32               { return logIdMint }
func (l *BurnLog) GetLogID() uint32               { return logIdBurn }
func (l *ReinitLog) GetLogID() uint32             { return logIdReinit }

type LogParser struct{}

func NewLogParser() (
	*LogParser,
	error,
) {
	return &LogParser{}, nil
}

func (c *LogParser) Parse(logCell *cell.Cell) (LogInterface, error) {
	logSlice := logCell.BeginParse()
	logId := logSlice.MustLoadUInt(32)

	switch logId {
	case logIdMint:
		mintLog, err := parseMintLog(logSlice)
		return mintLog, err
	case logIdBurn:
		burnLog, err := parseBurnLog(logSlice)
		return burnLog, err
	case logIdReinit:
		reinitLog, err := parseReinitLog(logSlice)
		return reinitLog, err
	default:
		return nil, fmt.Errorf("[LogParser] unknown log type with log id %x", logId)
	}
}

func parseMintLog(logSlice *cell.Slice) (*MintLog, error) {
	amount := logSlice.MustLoadBigCoins()
	receiver := logSlice.MustLoadAddr()
	bitcoinTxId, err := chainhash.NewHash(logSlice.MustLoadSlice(256))
	if err != nil {
		return nil, err
	}
	return &MintLog{
		amount, receiver, bitcoinTxId,
	}, nil
}

func parseBurnLog(logSlice *cell.Slice) (*BurnLog, error) {
	id := uint32(logSlice.MustLoadUInt(32))
	amount := logSlice.MustLoadBigCoins()
	bitcoinTxId, err := chainhash.NewHash(logSlice.MustLoadSlice(256))
	if err != nil {
		return nil, err
	}
	senderAddr := logSlice.MustLoadAddr()
	bitcoinScriptSlice := logSlice.MustLoadRef()
	bitcoinScript := bitcoinScriptSlice.MustLoadSlice(uint(bitcoinScriptSlice.MustLoadUInt(8) * 8))
	return &BurnLog{
		LogWithPegout{id, amount, bitcoinScript}, bitcoinTxId, senderAddr,
	}, nil
}

func parseReinitLog(logSlice *cell.Slice) (*ReinitLog, error) {
	id := uint32(logSlice.MustLoadUInt(32))
	amount := logSlice.MustLoadBigCoins()
	bitcoinTxId, err := chainhash.NewHash(logSlice.MustLoadSlice(256))
	if err != nil {
		return nil, err
	}
	bitcoinScriptSlice := logSlice.MustLoadRef()
	var bitcoinScript []byte
	if bitcoinScriptSlice.BitsLeft() > 2 {
		bitcoinScript = bitcoinScriptSlice.MustLoadSlice(bitcoinScriptSlice.BitsLeft())
	}
	return &ReinitLog{
		LogWithPegout{id, amount, bitcoinScript}, bitcoinTxId,
	}, nil
}
