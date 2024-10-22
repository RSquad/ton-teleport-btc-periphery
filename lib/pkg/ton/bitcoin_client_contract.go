package ton

import (
	"context"
	"math/rand/v2"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"lib/pkg/bitcoin"
)

const medianTimeSpan = 11
const blockHeaderBitLength = 80 * 8
const opCodeNewBlock = 0x5eefbc61

type BitcoinClientContract struct {
	Address *address.Address
	sender  *WalletContract
	api     *ton.APIClient
	ctx     context.Context
}

func NewBitcoinClientContract(
	api *ton.APIClient,
	address *address.Address,
	sender *WalletContract,
	ctx context.Context,
) *BitcoinClientContract {
	return &BitcoinClientContract{
		Address: address,
		sender:  sender,
		api:     api,
		ctx:     ctx,
	}
}

func (c *BitcoinClientContract) SendCandidateBlockHeader(candidateBlockHeader *wire.BlockHeader) (
	*tlb.Transaction,
	*ton.BlockIDExt,
	error,
) {
	queryID := rand.Uint64()
	candidateBlockHeaderBytes, err := bitcoin.BlockHeaderToBytes(candidateBlockHeader)
	if err != nil {
		return nil, nil, err
	}

	payload := cell.BeginCell().
		MustStoreUInt(opCodeNewBlock, 32).
		MustStoreUInt(queryID, 64).
		MustStoreRef(
			cell.BeginCell().
				MustStoreSlice(candidateBlockHeaderBytes, uint(len(candidateBlockHeaderBytes))*8).
				EndCell(),
		).EndCell()

	message := wallet.SimpleMessage(c.Address, tlb.MustFromTON("0.1"), payload)

	return c.sender.SendWaitTransaction(c.ctx, message)
}

func (c *BitcoinClientContract) GetStorageCell() (*cell.Cell, error) {
	block, err := c.api.CurrentMasterchainInfo(c.ctx)
	if err != nil {
		return nil, err
	}

	storage, err := c.api.RunGetMethod(c.ctx, block, c.Address, "get_data_tree")
	if err != nil {
		return nil, err
	}

	return storage.MustCell(0), nil
}

func (c *BitcoinClientContract) GetLastConfirmedBlockHash() (*chainhash.Hash, error) {
	storageCell, err := c.GetStorageCell()
	if err != nil {
		return nil, err
	}

	return chainhash.NewHash(storageCell.BeginParse().MustLoadSlice(256))
}

func (c *BitcoinClientContract) GetCandidateBlockHashes() ([]*chainhash.Hash, error) {
	storageCell, err := c.GetStorageCell()
	if err != nil {
		return nil, err
	}

	storageSlice := storageCell.BeginParse()
	storageSlice.MustLoadRef()
	storageSlice.MustLoadSlice(256 + 32 + 32 + 32 + 256 + 2 + 32*medianTimeSpan)

	var (
		blockHashes  []*chainhash.Hash
		blockHeaders []*cell.Slice
	)

	branch := storageSlice.MustLoadMaybeRef()
	for branch != nil {
		blockHeader := branch.MustLoadSlice(blockHeaderBitLength)

		blockHash := chainhash.DoubleHashH(blockHeader)

		blockHashes = append(blockHashes, &blockHash)
		blockHeaders = append(blockHeaders, branch)

		if storageSlice.BitsLeft() > 0 {
			branch = storageSlice.MustLoadMaybeRef()
		} else {
			branch = nil
		}

		if branch == nil && storageSlice.BitsLeft() > 0 {
			storageSlice = storageSlice.MustLoadRef()
			branch = storageSlice.MustLoadMaybeRef()
		}
	}

	for _, blockHeader := range blockHeaders {
		blockHeader.MustLoadUInt(32)
		blockHash, err := chainhash.NewHash(blockHeader.MustLoadSlice(256))
		if err != nil {
			panic(err)
		}

		if !bitcoin.SliceOfHashesContains(blockHashes, blockHash) {
			blockHashes = append(blockHashes, blockHash)
		}

		if blockHeader.BitsLeft() > 0 {
			blockHeaders = append(blockHeaders, blockHeader.MustLoadRef())
		}
	}

	return blockHashes, nil
}
