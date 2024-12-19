package mintmanager

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	ent "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
	internalkeymodel "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/internalkey"
	mintmodel "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/mint"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/pegincontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/toncenterv3"
	tonclient "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type MintManager struct {
	ctx               context.Context
	repo              *ent.Client
	bitcoinClient     *bitcoin.Client
	tonClient         *tonclient.TonClient
	tonCenterV3Client *toncenterv3.Client
	teleportContract  *teleportcontract.TeleportContract
}

func New(
	ctx context.Context,
	repo *ent.Client,
	bitcoinClient *bitcoin.Client,
	tonClient *tonclient.TonClient,
	tonCenterV3Client *toncenterv3.Client,
	teleportContract *teleportcontract.TeleportContract,
) *MintManager {
	return &MintManager{
		ctx:               ctx,
		repo:              repo,
		bitcoinClient:     bitcoinClient,
		tonClient:         tonClient,
		tonCenterV3Client: tonCenterV3Client,
		teleportContract:  teleportContract,
	}
}

func (c *MintManager) Run() {
	const sleepDuration = 3 * time.Second
	for {
		if err := c.processMints(); err != nil {
			log.Printf("failed to process pegins: %v", err)
		}
		time.Sleep(sleepDuration)
	}
}

func (c *MintManager) processMints() error {
	mints, err := c.repo.Mint.Query().
		Where(mintmodel.StatusNotIn(mintmodel.StatusSuccess, mintmodel.StatusRefunded)).
		Limit(128).
		WithPegin().
		All(c.ctx)
	if err != nil {
		return fmt.Errorf("failed to query mints: %w", err)
	}

	block, err := c.tonClient.API.CurrentMasterchainInfo(c.ctx)
	if err != nil {
		return fmt.Errorf("failed to get current masterchain info: %w", err)
	}

	teleportStorage, err := c.teleportContract.GetStorage(block)
	if err != nil {
		return fmt.Errorf("failed to get teleport contract storage: %w", err)
	}

	var wg sync.WaitGroup
	for _, mint := range mints {
		wg.Add(1)
		go func(mint *ent.Mint) {
			defer wg.Done()
			if err := c.processMint(mint, teleportStorage.PeginContractCode, block); err != nil {
				log.Printf("failed to process mint %v: %v", mint.ID, err)
			}
		}(mint)
	}

	wg.Wait()
	return nil
}

func (c *MintManager) processMint(
	mint *ent.Mint,
	peginContractCode *cell.Cell,
	block *ton.BlockIDExt,
) error {
	switch mint.Status {
	case mintmodel.StatusPending:
		return c.handlePendingMint(mint, peginContractCode, block)
	case mintmodel.StatusRefund:
		return c.handleRefundMint(mint)
	}
	return nil
}

func (c *MintManager) handlePendingMint(
	mint *ent.Mint,
	peginContractCode *cell.Cell,
	block *ton.BlockIDExt,
) error {
	log.Printf("BitcoinTxID %v", mint.Edges.Pegin.BitcoinTxID)
	bitcoinTxID, err := chainhash.NewHash(utils.MustHexToBytes(mint.Edges.Pegin.BitcoinTxID, 32))
	if err != nil {
		return fmt.Errorf("failed to create bitcoin transaction hash: %w", err)
	}

	peginContract, err := pegincontract.NewFromStateInit(
		&pegincontract.StateInit{
			Code: peginContractCode,
			InitData: &pegincontract.InitData{
				BitcoinTxID:          bitcoinTxID,
				TeleportContractAddr: c.teleportContract.Addr,
			},
		},
		c.tonClient,
		c.ctx,
	)
	if err != nil {
		return fmt.Errorf("failed to create pegin contract from state init: %w", err)
	}

	peginState, err := c.tonClient.API.GetAccount(c.ctx, block, peginContract.Addr)
	if err != nil {
		return fmt.Errorf("failed to get pegin contract account state: %w", err)
	}

	if peginState.IsActive {
		return c.updateMintStatus(mint.ID, mintmodel.StatusSuccess)
	}

	internalKey, err := c.repo.InternalKey.Query().
		Where(internalkeymodel.KeyEQ(mint.Edges.Pegin.InternalKey)).
		Only(c.ctx)
	if err != nil {
		return fmt.Errorf("failed to query internal key: %w", err)
	}

	latestInternalKey, err := c.repo.InternalKey.Query().
		Order(ent.Desc(internalkeymodel.FieldCompletedAt)).
		First(c.ctx)
	if err != nil {
		return fmt.Errorf("failed to query latest internal key: %w", err)
	}

	if !internalKey.CompletedAt.Equal(latestInternalKey.CompletedAt) {
		return c.updateMintStatus(mint.ID, mintmodel.StatusRefund)
	}

	return nil
}

func (c *MintManager) handleRefundMint(mint *ent.Mint) error {
	bitcoinTxID, err := chainhash.NewHashFromStr(mint.Edges.Pegin.BitcoinTxID)
	if err != nil {
		return fmt.Errorf("invalid bitcoin tx id: %w", err)
	}

	out, err := c.bitcoinClient.RPCClient.GetTxOut(bitcoinTxID, uint32(mint.Edges.Pegin.VoutIndex), false)
	if err != nil {
		return fmt.Errorf("failed to get bitcoin transaction output: %w", err)
	}

	if out == nil {
		return c.updateMintStatus(mint.ID, mintmodel.StatusRefunded)
	}

	return nil
}

func (c *MintManager) updateMintStatus(mintID int, status mintmodel.Status) error {
	err := c.repo.Mint.Update().
		SetStatus(status).
		Where(mintmodel.ID(mintID)).
		Exec(c.ctx)
	if err != nil {
		log.Printf("failed to update mint status for %v: %v", mintID, err)
	}
	return err
}
