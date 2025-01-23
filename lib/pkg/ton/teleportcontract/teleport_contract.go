package teleportcontract

import (
	"bytes"
	"context"
	"encoding/binary"
	"math/big"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	tonutils "github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	jwv4r2contract "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/jw_v4r2_contract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
)

const (
	opCodeConfirmPegoutTx          = 0xbd0eaf09
	storageIndexPegoutChainCounter = 12
	storageIndexLastPegoutTxID     = 13
	storageIndexCsvLock            = 15
	storageIndexLimits             = 16
	storageIndexPegoutContractCode = 6
	storageIndexPeginContractCode  = 9
)

type TeleportContract struct {
	ton.Contract
	TonClient *tonclient.TonClient
	sender    *jwv4r2contract.JWV4R2Contract
	ctx       context.Context
}

type Storage struct {
	PegoutContractCode *cell.Cell
	PegoutChainCounter uint64
	LastPegoutTxID     *chainhash.Hash
	CsvLock            uint32
	PeginContractCode  *cell.Cell
	Limits             Limits
}

type Limits struct {
	MinPeginAmount  uint32
	MinPegoutAmount uint32
}

func New(
	addr *address.Address,
	tonClient *tonclient.TonClient,
	sender *jwv4r2contract.JWV4R2Contract,
	ctx context.Context,
) *TeleportContract {
	return &TeleportContract{ton.Contract{Addr: addr}, tonClient, sender, ctx}
}

func (c *TeleportContract) SendPegoutProof(
	txID *chainhash.Hash,
	blockHash *chainhash.Hash,
	merkleBlock *wire.MsgMerkleBlock,
) (
	*tlb.Transaction,
	*tonutils.BlockIDExt,
	error,
) {
	blockHashUInt := new(big.Int).SetBytes(blockHash.CloneBytes())
	txIDUInt := new(big.Int).SetBytes(txID.CloneBytes())

	const txCountLen = 4
	txCount := make([]byte, txCountLen)
	binary.LittleEndian.PutUint32(txCount, merkleBlock.Transactions)

	hashesCountBuf := new(bytes.Buffer)
	if err := wire.WriteVarInt(hashesCountBuf, 0, uint64(len(merkleBlock.Hashes))); err != nil {
		return nil, nil, err
	}
	hashesCount := hashesCountBuf.Bytes()

	hashesBuilder := cell.BeginCell()
	c.storeHashesToCell(merkleBlock.Hashes, hashesBuilder)
	hashesCell := hashesBuilder.EndCell()

	flagsLenBuf := new(bytes.Buffer)
	if err := wire.WriteVarInt(flagsLenBuf, 0, uint64(len(merkleBlock.Flags))); err != nil {
		return nil, nil, err
	}
	flagsLen := flagsLenBuf.Bytes()

	flagsCell := cell.BeginCell().MustStoreBinarySnake(merkleBlock.Flags).EndCell()

	proofCell := cell.BeginCell().
		MustStoreSlice(txCount, txCountLen*8).
		MustStoreSlice(hashesCount, uint(len(hashesCount))*8).
		MustStoreRef(hashesCell).
		MustStoreSlice(flagsLen, uint(len(flagsLen))*8).
		MustStoreRef(flagsCell).EndCell()

	payload := cell.BeginCell().
		MustStoreUInt(opCodeConfirmPegoutTx, 32).
		MustStoreBigUInt(blockHashUInt, 256).
		MustStoreBigUInt(txIDUInt, 256).MustStoreRef(proofCell).EndCell()

	message := wallet.SimpleMessage(c.Addr, tlb.MustFromTON("0.1"), payload)

	return c.sender.SendWaitTransaction(c.ctx, message)
}

func (c *TeleportContract) GetStorage(block *tonutils.BlockIDExt) (Storage, error) {
	if block == nil {
		var err error
		block, err = c.TonClient.API.CurrentMasterchainInfo(c.ctx)
		if err != nil {
			return Storage{}, err
		}
	}

	storage, err := c.TonClient.API.RunGetMethod(c.ctx, block, c.Addr, "get_storage")
	if err != nil {
		return Storage{}, err
	}

	pegoutChainCounter := storage.MustInt(storageIndexPegoutChainCounter)
	lastPegoutTxIDInt := storage.MustInt(storageIndexLastPegoutTxID)
	pegoutContractCode := storage.MustCell(storageIndexPegoutContractCode)
	peginContractCode := storage.MustCell(storageIndexPeginContractCode)
	csvLock := storage.MustInt(storageIndexCsvLock)
	limitsSlice := storage.MustSlice(storageIndexLimits)
	limits := Limits{
		MinPeginAmount:  uint32(limitsSlice.MustLoadUInt(32)),
		MinPegoutAmount: uint32(limitsSlice.MustLoadUInt(32)),
	}
	lastPegoutTxID, err := chainhash.NewHash(utils.BytesPadTo(lastPegoutTxIDInt.Bytes(), 32))
	if err != nil {
		return Storage{}, err
	}

	return Storage{
		PegoutContractCode: pegoutContractCode,
		PegoutChainCounter: pegoutChainCounter.Uint64(),
		LastPegoutTxID:     lastPegoutTxID,
		CsvLock:            uint32(csvLock.Uint64()),
		PeginContractCode:  peginContractCode,
		Limits:             limits,
	}, nil
}

func (c *TeleportContract) storeHashesToCell(
	hashes []*chainhash.Hash,
	builder *cell.Builder,
) *cell.Builder {
	const hashBitLen = 256
	var store func(hashes []*chainhash.Hash, builder *cell.Builder) *cell.Builder
	store = func(hashes []*chainhash.Hash, builder *cell.Builder) *cell.Builder {
		if len(hashes) == 0 {
			return builder
		}

		space := int(builder.BitsLeft() / hashBitLen)
		n := int(utils.MinInt(int64(space), int64(len(hashes))))

		for i := 0; i < space && i < len(hashes); i++ {
			builder.MustStoreSlice(hashes[i].CloneBytes(), hashBitLen)
		}

		if n < len(hashes) {
			refCell := cell.BeginCell()
			store(hashes[n:], refCell)
			builder.MustStoreRef(refCell.EndCell())
		}

		return builder
	}

	return store(hashes, builder)
}
