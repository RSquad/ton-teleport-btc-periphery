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

type MintLog struct {
	Amount       *big.Int
	ReceiverAddr *address.Address
	BitcoinTxID  *chainhash.Hash
}

func (m *MintLog) GetLogID() uint32 {
	return logIdMint
}

type BurnLog struct {
	ID            uint32
	Amount        *big.Int
	SenderAddr    *address.Address
	BitcoinTxID   *chainhash.Hash
	BitcoinScript []byte
}

func (b *BurnLog) GetLogID() uint32 {
	return logIdBurn
}

type ReinitLog struct {
	ID            uint32
	Amount        *big.Int
	BitcoinTxID   *chainhash.Hash
	BitcoinScript []byte
}

func (r *ReinitLog) GetLogID() uint32 {
	return logIdReinit
}

type LogParser struct{}

func NewTeleportContractLogParser() (
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
	bitcoinTxID, err := chainhash.NewHash(logSlice.MustLoadSlice(256))
	if err != nil {
		return nil, err
	}
	return &MintLog{
		amount, receiver, bitcoinTxID,
	}, nil
}

func parseBurnLog(logSlice *cell.Slice) (*BurnLog, error) {
	id := uint32(logSlice.MustLoadUInt(32))
	amount := logSlice.MustLoadBigCoins()
	bitcoinTxID, err := chainhash.NewHash(logSlice.MustLoadSlice(256))
	if err != nil {
		return nil, err
	}
	senderAddr := logSlice.MustLoadAddr()
	bitcoinScriptSlice := logSlice.MustLoadRef()
	bitcoinScript := bitcoinScriptSlice.MustLoadSlice(uint(bitcoinScriptSlice.MustLoadUInt(8)))
	return &BurnLog{
		id, amount, senderAddr, bitcoinTxID, bitcoinScript,
	}, nil
}

func parseReinitLog(logSlice *cell.Slice) (*ReinitLog, error) {
	id := uint32(logSlice.MustLoadUInt(32))
	amount := logSlice.MustLoadBigCoins()
	bitcoinTxID, err := chainhash.NewHash(logSlice.MustLoadSlice(256))
	if err != nil {
		return nil, err
	}
	bitcoinScriptSlice := logSlice.MustLoadRef()
	var bitcoinScript []byte
	if bitcoinScriptSlice.BitsLeft() > 2 {
		bitcoinScript = bitcoinScriptSlice.MustLoadSlice(bitcoinScriptSlice.BitsLeft())
	}
	return &ReinitLog{
		id, amount, bitcoinTxID, bitcoinScript,
	}, nil
}
