package bitcoinclientcontract

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

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	jwv4r2contract "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/jw_v4r2_contract"
	tonclient "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
)

const (
	medianTimeSpan       = 11
	blockHeaderBitLength = 80 * 8
	opCodeNewBlock       = 0x5eefbc61
)

type BitcoinClientContract struct {
	Addr      *address.Address
	tonClient *tonclient.TonClient
	sender    *jwv4r2contract.JWV4R2Contract
	ctx       context.Context
}

func NewBitcoinClientContract(
	addr *address.Address,
	tonClient *tonclient.TonClient,
	sender *jwv4r2contract.JWV4R2Contract,
	ctx context.Context,
) *BitcoinClientContract {
	return &BitcoinClientContract{
		addr, tonClient, sender, ctx,
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

	message := wallet.SimpleMessage(c.Addr, tlb.MustFromTON("0.1"), payload)

	return c.sender.SendWaitTransaction(c.ctx, message)
}

func (c *BitcoinClientContract) GetStorageCell() (*cell.Cell, error) {
	block, err := c.tonClient.API.CurrentMasterchainInfo(c.ctx)
	if err != nil {
		return nil, err
	}

	storage, err := c.tonClient.API.RunGetMethod(c.ctx, block, c.Addr, "get_data_tree")
	if err != nil {
		return nil, err
	}

	return storage.MustCell(0), nil
}

func (c *BitcoinClientContract) GetConfirmationsNeeded() (int64, error) {
	storageCell, err := c.GetStorageCell()
	if err != nil {
		return 0, err
	}

	return c.GetConfirmationsNeededFromCell(storageCell), nil
}

func (c *BitcoinClientContract) GetConfirmationsNeededFromCell(storageCell *cell.Cell) int64 {
	storageSlice := storageCell.BeginParse()
	storageSlice.MustLoadRef()
	storageSlice.MustLoadSlice(256)
	storageSlice.MustLoadBigUInt(32)

	return storageSlice.MustLoadInt(4)
}

func (c *BitcoinClientContract) GetLastConfirmedBlockHash() (*chainhash.Hash, error) {
	storageCell, err := c.GetStorageCell()
	if err != nil {
		return nil, err
	}

	return c.GetLastConfirmedBlockHashFromCell(storageCell)
}

func (c *BitcoinClientContract) GetLastConfirmedBlockHashFromCell(storageCell *cell.Cell) (*chainhash.Hash, error) {
	return chainhash.NewHash(storageCell.BeginParse().MustLoadSlice(256))
}

func (c *BitcoinClientContract) GetCandidateBlockHashes() ([]*chainhash.Hash, error) {
	storageCell, err := c.GetStorageCell()
	if err != nil {
		return nil, err
	}

	return c.GetCandidateBlockHashesFromCell(storageCell)
}

func (c *BitcoinClientContract) GetCandidateBlockHashesFromCell(storageCell *cell.Cell) ([]*chainhash.Hash, error) {
	storageSlice := storageCell.BeginParse()
	storageSlice.MustLoadRef()
	storageSlice.MustLoadSlice(256 + 32 + 4 + 32 + 32 + 256 + 2 + 32*medianTimeSpan)

	var (
		blockHashes  []*chainhash.Hash
		blockHeaders []*cell.Slice
	)

	branch := storageSlice.MustLoadMaybeRef()
	for branch != nil {
		branchCopy := *branch
		blockHash := c.CalcBlockHashFromSlice(&branchCopy)

		blockHashes = append(blockHashes, blockHash)
		blockHeaders = append(blockHeaders, branch)

		if storageSlice.BitsLeft() > 0 {
			branch = storageSlice.MustLoadMaybeRef()
		} else {
			branch = nil
		}

		if branch == nil && storageSlice.RefsNum() > 0 {
			storageSlice = storageSlice.MustLoadRef()
			branch = storageSlice.MustLoadMaybeRef()
		}
	}

	for i := range blockHeaders {
		blockHeader := blockHeaders[i]
		blockHeader.MustLoadUInt(32)
		blockHash, err := chainhash.NewHash(blockHeader.MustLoadSlice(256))
		if err != nil {
			return nil, err
		}

		if !bitcoin.SliceOfHashesContains(blockHashes, blockHash) {
			blockHashes = append(blockHashes, blockHash)
		}

		if blockHeader.RefsNum() > 0 {
			blockHeaders = append(blockHeaders, blockHeader.MustLoadRef())
		}
	}

	return blockHashes, nil
}

func (c *BitcoinClientContract) CalcBlockHashFromSlice(blockHeaderCell *cell.Slice) *chainhash.Hash {
	blockHeader := blockHeaderCell.MustLoadSlice(blockHeaderBitLength)

	hash := chainhash.DoubleHashH(blockHeader)
	return &hash
}
