package ton

import (
    "bytes"
    "context"
    "encoding/binary"
    "math/big"
    "math/rand/v2"

    "github.com/btcsuite/btcd/chaincfg/chainhash"
    "github.com/btcsuite/btcd/wire"
    "github.com/xssnick/tonutils-go/address"
    "github.com/xssnick/tonutils-go/tlb"
    "github.com/xssnick/tonutils-go/ton"
    "github.com/xssnick/tonutils-go/ton/wallet"
    "github.com/xssnick/tonutils-go/tvm/cell"

    "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
)

const opCodeConfirmPegoutTx = 0xbd0eaf09
const storageIndexPegoutChainCounter = 13
const storageIndexLastPegoutTxID = 14

type TeleportContract struct {
    Address *address.Address
    sender  *WalletContract
    api     *ton.APIClient
    ctx     context.Context
}

type TeleportContractStorage struct {
    PegoutChainCounter uint64
    LastPegoutTxID     *chainhash.Hash
}

func NewTeleportContract(
    api *ton.APIClient,
    address *address.Address,
    sender *WalletContract,
    ctx context.Context,
) *TeleportContract {
    return &TeleportContract{
        Address: address,
        sender:  sender,
        api:     api,
        ctx:     ctx,
    }
}

func (c *TeleportContract) SendPegoutProof(
    txID *chainhash.Hash,
    blockHash *chainhash.Hash,
    merkleBlock *wire.MsgMerkleBlock,
) (
    *tlb.Transaction,
    *ton.BlockIDExt,
    error,
) {
    queryID := rand.Uint64()

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
        MustStoreUInt(queryID, 64).
        MustStoreBigUInt(blockHashUInt, 256).
        MustStoreBigUInt(txIDUInt, 256).MustStoreRef(proofCell).EndCell()

    message := wallet.SimpleMessage(c.Address, tlb.MustFromTON("0.1"), payload)

    return c.sender.SendWaitTransaction(c.ctx, message)
}

func (c *TeleportContract) GetStorage() (TeleportContractStorage, error) {
    block, err := c.api.CurrentMasterchainInfo(c.ctx)
    if err != nil {
        return TeleportContractStorage{}, err
    }

    storage, err := c.api.RunGetMethod(c.ctx, block, c.Address, "get_storage")
    if err != nil {
        return TeleportContractStorage{}, err
    }

    pegoutChainCounter := storage.MustInt(storageIndexPegoutChainCounter)

    lastPegoutTxIDInt := storage.MustInt(storageIndexLastPegoutTxID)

    lastPegoutTxID, err := chainhash.NewHash(lastPegoutTxIDInt.Bytes())
    if err != nil {
        return TeleportContractStorage{}, err
    }

    return TeleportContractStorage{
        PegoutChainCounter: pegoutChainCounter.Uint64(),
        LastPegoutTxID:     lastPegoutTxID,
    }, nil
}

func (c *TeleportContract) storeHashesToCell(hashes []*chainhash.Hash, builder *cell.Builder) *cell.Builder {
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
