package teleportcontract

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand/v2"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	tonutils "github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	jwv4r2contract "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/jw_v4r2_contract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/parseddict"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
)

const (
	opCodeConfirmPegoutTx            = 0xbd0eaf09
	opCodeTeleportTransferBtc        = 0x3f781d24
	storageIndexId                   = 0
	storageIndexBlockCode            = 1
	storageIndexDeposits             = 2
	storageIndexMinterAddress        = 3
	storageIndexBitcoinClientAddress = 4
	storageIndexTweakedPubkey        = 5
	storageIndexPegoutContractCode   = 6
	storageIndexCoordinatorAddress   = 7
	storageIndexUTXOset              = 8
	storageIndexPeginContractCode    = 9
	storageIndexNextSVB              = 10
	storageIndexBaseSVB              = 11
	storageIndexPegoutChainCounter   = 12
	storageIndexLastPegoutTxID       = 13
	storageIndexInternalKey          = 14
	storageIndexCsvLock              = 15
	storageIndexLimits               = 16
	storageIndexConfiguratorAddress  = 17
	storageIndexTotalServiceFee      = 18
	storageIndexEnabled              = 19
	storageIndexInspectorAddress     = 20
)

type TeleportContract struct {
	ton.Contract
	TonClient *tonclient.TonClient
	sender    *jwv4r2contract.JWV4R2Contract
	ctx       context.Context
}

type Storage struct {
	Id                   uint16
	BlockCode            *cell.Cell
	PegoutContractCode   *cell.Cell
	PeginContractCode    *cell.Cell
	MinterAddress        *address.Address
	BitcoinClientAddress *address.Address
	CoordinatorAddress   *address.Address
	InspectorAddress     *address.Address
	ConfiguratorAddress  *address.Address
	TweakedPubkey        string
	InternalKey          string
	Deposits             map[uint64]DepositData
	NextSVB              uint16
	BaseSVB              uint16
	PegoutChainCounter   uint64
	LastPegoutTxID       *chainhash.Hash
	CsvLock              uint32
	Limits               Limits
	TotalServiceFee      int32
	Enabled              bool
	UTXOset              map[string]UTXOData
}

type DepositData struct {
	BlockHash       *chainhash.Hash
	TxID            *chainhash.Hash
	DestAddress     *address.Address
	ResponseAddress *address.Address
	Amount          *big.Int
	TapMerkleRoot   *chainhash.Hash
	TxProof         []byte
	Tx              []byte
}

type UTXOData struct {
	Amount        *big.Int
	Index         uint8
	TapMerkleRoot *chainhash.Hash
	MintAddress   *address.Address
	Script        string
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

	proofCell, err := c.buildProofCell(merkleBlock)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build proof cell: %w", err)
	}

	payload := cell.BeginCell().
		MustStoreUInt(opCodeConfirmPegoutTx, 32).
		MustStoreBigUInt(blockHashUInt, 256).
		MustStoreBigUInt(txIDUInt, 256).MustStoreRef(proofCell).EndCell()

	message := wallet.SimpleMessage(c.Addr, tlb.MustFromTON("0.1"), payload)

	return c.sender.SendWaitTransaction(c.ctx, message)
}

func (c *TeleportContract) CheckContractBalance(ctx context.Context) error {
	block, err := c.TonClient.API.CurrentMasterchainInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current masterchain info: %w", err)
	}

	acc, err := c.TonClient.API.GetAccount(ctx, block, c.Addr)
	if err != nil {
		return fmt.Errorf("GetAccount: %w", err)
	}

	balanceNano := acc.State.Balance.Nano()

	TONInNano := int64(1_000_000_000)

	if balanceNano.Int64() < TONInNano {
		return fmt.Errorf("not enough balance")
	}

	return nil
}

func (c *TeleportContract) SendDeposit(
	ctx context.Context,
	blockHash *chainhash.Hash,
	receiverAddressStr string,
	indexerAddressStr string,
	recoveryKey string,
	txHex string,
	txProof []byte,
) (*tlb.Transaction, *tonutils.BlockIDExt, error) {
	if err := c.CheckContractBalance(ctx); err != nil {
		return nil, nil, fmt.Errorf("not enough money to send deposit: %w", err)
	}

	blockHashUInt := new(big.Int).SetBytes(blockHash.CloneBytes())
	destAddress := address.MustParseRawAddr(receiverAddressStr)
	indexerAddress := address.MustParseAddr(indexerAddressStr)
	queryId := rand.Uint64()

	decodedTxProof, err := c.decodeTxProof(txProof)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode tx proof: %w", err)
	}

	proofCell, err := c.buildProofCell(decodedTxProof)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build proof cell: %w", err)
	}

	recoveryKeyBytes, err := hex.DecodeString(recoveryKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode recovery key: %w", err)
	}

	serializedTransaction, err := c.serializeTransaction(txHex)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	sendDepositBodyCell := cell.BeginCell().
		MustStoreUInt(opCodeTeleportTransferBtc, 32).
		MustStoreUInt(queryId, 64).
		MustStoreSlice(blockHashUInt.Bytes(), 256).
		MustStoreRef(serializedTransaction).
		MustStoreRef(proofCell).
		MustStoreMaybeRef(cell.BeginCell().MustStoreBinarySnake(recoveryKeyBytes).EndCell()).
		MustStoreAddr(destAddress).
		MustStoreAddr(indexerAddress).
		EndCell()

	message := wallet.SimpleMessage(c.Addr, tlb.MustFromTON("1"), sendDepositBodyCell)

	return c.sender.SendWaitTransaction(c.ctx, message)
}

func (c *TeleportContract) GetAutoPegoutFee(block *tonutils.BlockIDExt) (*big.Int, error) {
	if block == nil {
		var err error
		block, err = c.TonClient.API.CurrentMasterchainInfo(c.ctx)
		if err != nil {
			return nil, err
		}
	}
	result, err := c.TonClient.API.RunGetMethod(c.ctx, block, c.Addr, "get_auto_pegout_fee")

	if err != nil {
		return nil, err
	}
	return result.MustInt(0), nil
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

	id := uint16(storage.MustInt(storageIndexId).Uint64())
	blockCode := storage.MustCell(storageIndexBlockCode)
	pegoutContractCode := storage.MustCell(storageIndexPegoutContractCode)
	peginContractCode := storage.MustCell(storageIndexPeginContractCode)

	minterAddress := storage.MustSlice(storageIndexMinterAddress).MustLoadAddr()
	bitcoinClientAddress := storage.MustSlice(storageIndexBitcoinClientAddress).MustLoadAddr()
	coordinatorAddress := storage.MustSlice(storageIndexCoordinatorAddress).MustLoadAddr()
	inspectorAddress := storage.MustSlice(storageIndexInspectorAddress).MustLoadAddr()
	configuratorAddress := storage.MustSlice(storageIndexConfiguratorAddress).MustLoadAddr()

	tweakedPubkey := hex.EncodeToString(utils.BytesPadTo(storage.MustInt(storageIndexTweakedPubkey).Bytes(), 32))
	internalKey := hex.EncodeToString(utils.BytesPadTo(storage.MustInt(storageIndexInternalKey).Bytes(), 32))

	nextSVB := uint16(storage.MustInt(storageIndexNextSVB).Uint64())
	baseSVB := uint16(storage.MustInt(storageIndexBaseSVB).Uint64())
	pegoutChainCounter := storage.MustInt(storageIndexPegoutChainCounter).Uint64()
	lastPegoutTxID, err := chainhash.NewHash(utils.BytesPadTo(storage.MustInt(storageIndexLastPegoutTxID).Bytes(), 32))
	if err != nil {
		return Storage{}, err
	}

	csvLock := uint32(storage.MustInt(storageIndexCsvLock).Uint64())
	limitsSlice := storage.MustSlice(storageIndexLimits)
	limits := Limits{
		MinPeginAmount:  uint32(limitsSlice.MustLoadUInt(32)),
		MinPegoutAmount: uint32(limitsSlice.MustLoadUInt(32)),
	}

	totalServiceFee := int32(storage.MustInt(storageIndexTotalServiceFee).Int64())
	enabled := storage.MustInt(storageIndexEnabled).Bit(0) > 0

	// Deposits
	depositsCell, _ := storage.Cell(storageIndexDeposits)
	deposits := map[uint64]DepositData{}
	if depositsCell != nil {
		deposits, err = loadDepositsMap(depositsCell)
		if err != nil {
			return Storage{}, err
		}
	}

	// UTXO set
	utxoSetCell, _ := storage.Cell(storageIndexUTXOset)
	utxoSet := map[string]UTXOData{}
	if utxoSetCell != nil {
		utxoSet, err = loadUTXOset(utxoSetCell)
		if err != nil {
			return Storage{}, err
		}
	}

	return Storage{
		Id:                   id,
		BlockCode:            blockCode,
		PegoutContractCode:   pegoutContractCode,
		PeginContractCode:    peginContractCode,
		MinterAddress:        minterAddress,
		BitcoinClientAddress: bitcoinClientAddress,
		CoordinatorAddress:   coordinatorAddress,
		InspectorAddress:     inspectorAddress,
		ConfiguratorAddress:  configuratorAddress,
		TweakedPubkey:        tweakedPubkey,
		InternalKey:          internalKey,
		Deposits:             deposits,
		NextSVB:              nextSVB,
		BaseSVB:              baseSVB,
		PegoutChainCounter:   pegoutChainCounter,
		LastPegoutTxID:       lastPegoutTxID,
		CsvLock:              csvLock,
		Limits:               limits,
		TotalServiceFee:      totalServiceFee,
		Enabled:              enabled,
		UTXOset:              utxoSet,
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

func loadDepositsMap(depositsCell *cell.Cell) (map[uint64]DepositData, error) {
	deposits, err := parseddict.ParseDict(
		depositsCell.AsDict(64),
		parseddict.ParseKeyUI64,
		func(s *cell.Slice) (DepositData, error) {
			s = s.MustLoadRef()

			blockHash, err := chainhash.NewHash(s.MustLoadSlice(256))
			if err != nil {
				return DepositData{}, err
			}

			tapMerkleRoot, err := chainhash.NewHash(s.MustLoadSlice(256))
			if err != nil {
				return DepositData{}, err
			}

			txID, err := chainhash.NewHash(s.MustLoadSlice(256))
			if err != nil {
				return DepositData{}, err
			}
			amount := s.MustLoadBigUInt(128)
			txProofSlice := s.MustLoadRef()
			txProof := txProofSlice.MustLoadSlice(txProofSlice.BitsLeft())
			txSlice := s.MustLoadRef()
			tx := txSlice.MustLoadSlice(txSlice.BitsLeft())
			addressesSlice := s.MustLoadRef()
			destAddress := addressesSlice.MustLoadAddr()
			responseAddress := addressesSlice.MustLoadAddr()

			return DepositData{
				BlockHash:       blockHash,
				TxID:            txID,
				DestAddress:     destAddress,
				ResponseAddress: responseAddress,
				Amount:          amount,
				TapMerkleRoot:   tapMerkleRoot,
				TxProof:         txProof,
				Tx:              tx,
			}, nil
		},
	)
	if err != nil {
		return nil, err
	}

	return *deposits, nil
}

func loadUTXOset(utxoSetCell *cell.Cell) (map[string]UTXOData, error) {
	utxoSet, err := parseddict.ParseDict(
		utxoSetCell.AsDict(256),
		parseddict.ParseKeyBitcoinHashAsStr,
		func(s *cell.Slice) (UTXOData, error) {
			amount := s.MustLoadBigUInt(128)
			index := uint8(s.MustLoadUInt(8))

			tapMerkleRoot, err := chainhash.NewHash(s.MustLoadSlice(256))
			if err != nil {
				return UTXOData{}, err
			}

			mintAddress := s.MustLoadAddr()

			scriptSlice := s.MustLoadRef()
			script := hex.EncodeToString(scriptSlice.MustLoadSlice(scriptSlice.BitsLeft()))

			return UTXOData{
				Amount:        amount,
				Index:         index,
				TapMerkleRoot: tapMerkleRoot,
				MintAddress:   mintAddress,
				Script:        script,
			}, nil
		},
	)
	if err != nil {
		return nil, err
	}

	return *utxoSet, nil
}

func (c *TeleportContract) decodeTxProof(txProof []byte) (*wire.MsgMerkleBlock, error) {
	var merkleBlock wire.MsgMerkleBlock
	buf := bytes.NewBuffer(txProof)
	err := merkleBlock.BtcDecode(buf, wire.ProtocolVersion, wire.BaseEncoding)
	if err != nil {
		return nil, err
	}

	return &merkleBlock, nil
}

func (c *TeleportContract) buildProofCell(merkleBlock *wire.MsgMerkleBlock) (*cell.Cell, error) {
	const txCountLen = 4
	txCount := make([]byte, txCountLen)
	binary.LittleEndian.PutUint32(txCount, merkleBlock.Transactions)

	hashesCountBuf := new(bytes.Buffer)
	if err := wire.WriteVarInt(hashesCountBuf, 0, uint64(len(merkleBlock.Hashes))); err != nil {
		return nil, fmt.Errorf("failed to write hashes count: %w", err)
	}
	hashesCount := hashesCountBuf.Bytes()

	hashesBuilder := cell.BeginCell()
	c.storeHashesToCell(merkleBlock.Hashes, hashesBuilder)
	hashesCell := hashesBuilder.EndCell()

	flagsLenBuf := new(bytes.Buffer)
	if err := wire.WriteVarInt(flagsLenBuf, 0, uint64(len(merkleBlock.Flags))); err != nil {
		return nil, fmt.Errorf("failed to write flags count: %w", err)
	}
	flagsLen := flagsLenBuf.Bytes()

	flagsCell := cell.BeginCell().MustStoreBinarySnake(merkleBlock.Flags).EndCell()

	proofCell := cell.BeginCell().
		MustStoreSlice(txCount, txCountLen*8).
		MustStoreSlice(hashesCount, uint(len(hashesCount))*8).
		MustStoreRef(hashesCell).
		MustStoreSlice(flagsLen, uint(len(flagsLen))*8).
		MustStoreRef(flagsCell).EndCell()

	return proofCell, nil
}
