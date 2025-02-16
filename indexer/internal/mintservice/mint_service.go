package mintservice

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	ent "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
	internalkeymodel "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/internalkey"
	mintmodel "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/mint"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/pegincontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	tonclient "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type MintService struct {
	repo             *ent.Client
	bitcoinClient    *bitcoin.Client
	tonClient        *tonclient.TonClient
	teleportContract *teleportcontract.TeleportContract
}

func New(
	repo *ent.Client,
	bitcoinClient *bitcoin.Client,
	tonClient *tonclient.TonClient,
	teleportContract *teleportcontract.TeleportContract,
) *MintService {
	return &MintService{
		repo:             repo,
		bitcoinClient:    bitcoinClient,
		tonClient:        tonClient,
		teleportContract: teleportContract,
	}
}

func (ms *MintService) Work(ctx context.Context) (err error) {
	defer ms.logFinishWork(err)
	ms.logStartWork()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			ms.processMints(ctx)
			time.Sleep(3 * time.Second)
		}
	}
}

func (ms *MintService) processMints(ctx context.Context) (err error) {
	start := time.Now()
	defer func() {
		logFinishProcessingMints(time.Since(start), err)
	}()
	logStartProcessingMints()

	mints, err := ms.queryUnprocessedMints(ctx)
	if err != nil {
		return fmt.Errorf(errQueryUnprocessedMints, err)
	}
	if len(mints) == 0 {
		logNoUnprocessedMints()
		return nil
	}
	logUnprocessedMintsReceived(len(mints))

	block, err := ms.tonClient.API.CurrentMasterchainInfo(ctx)
	if err != nil {
		return fmt.Errorf(tonclient.ErrGetCurrentMasterchainInfo, err)
	}

	teleportStorage, err := ms.teleportContract.GetStorage(block)
	if err != nil {
		return fmt.Errorf(teleportcontract.ErrGetStorage, err)
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 128)

	for i, mint := range mints {
		wg.Add(1)
		sem <- struct{}{}
		go func(mint *ent.Mint) {
			defer wg.Done()
			defer func() { <-sem }()
			ms.processMint(ctx, mint, teleportStorage.PeginContractCode, block)
			logMintsProcessingProgress(i+1, len(mints))
		}(mint)
	}

	wg.Wait()
	return nil
}

func (ms *MintService) processMint(
	ctx context.Context,
	mint *ent.Mint,
	peginContractCode *cell.Cell,
	block *ton.BlockIDExt,
) (err error) {
	switch mint.Status {
	case mintmodel.StatusPending:
		return ms.handlePendingMint(ctx, mint, peginContractCode, block)
	case mintmodel.StatusRefund:
		return ms.handleRefundMint(ctx, mint)
	}
	return nil
}

func (ms *MintService) handlePendingMint(
	ctx context.Context,
	mint *ent.Mint,
	peginContractCode *cell.Cell,
	block *ton.BlockIDExt,
) error {
	bitcoinTxID, err := chainhash.NewHash(utils.MustHexToBytes(mint.Edges.Pegin.BitcoinTxID, 32))
	if err != nil {
		return fmt.Errorf(errCalcBitcoinTxID, err)
	}

	peginContract, err := pegincontract.NewFromStateInit(
		ctx,
		&pegincontract.StateInit{
			Code: peginContractCode,
			InitData: &pegincontract.InitData{
				BitcoinTxID:          bitcoinTxID,
				TeleportContractAddr: ms.teleportContract.Addr,
			},
		},
		ms.tonClient,
	)
	if err != nil {
		return fmt.Errorf(errCreatePeginContractFromStateInit, err)
	}

	peginState, err := ms.tonClient.API.GetAccount(ctx, block, peginContract.Addr)
	if err != nil {
		return fmt.Errorf(errGetPeginContractAccountState, err)
	}

	if peginState.IsActive {
		return ms.updateMintStatus(ctx, mint.ID, mintmodel.StatusSuccess)
	}

	internalKey, err := ms.repo.InternalKey.Query().
		Where(internalkeymodel.KeyEQ(mint.Edges.Pegin.InternalKey)).
		Only(ctx)
	if err != nil {
		return fmt.Errorf(errQueryInternalKey, err)
	}

	latestInternalKey, err := ms.repo.InternalKey.Query().
		Order(ent.Desc(internalkeymodel.FieldCompletedAt)).
		First(ctx)
	if err != nil {
		return fmt.Errorf(errQueryLatestInternalKey, err)
	}

	if !internalKey.CompletedAt.Equal(latestInternalKey.CompletedAt) {
		return ms.updateMintStatus(ctx, mint.ID, mintmodel.StatusRefund)
	}

	return nil
}

func (ms *MintService) handleRefundMint(ctx context.Context, mint *ent.Mint) error {
	bitcoinTxID, err := chainhash.NewHashFromStr(mint.Edges.Pegin.BitcoinTxID)
	if err != nil {
		return err
	}

	out, err := ms.bitcoinClient.RPCClient.GetTxOut(bitcoinTxID, uint32(mint.Edges.Pegin.VoutIndex), true)
	if err != nil {
		return err
	}

	if out == nil {
		return ms.updateMintStatus(ctx, mint.ID, mintmodel.StatusRefunded)
	}

	return nil
}

func (ms *MintService) updateMintStatus(ctx context.Context, mintID int, status mintmodel.Status) error {
	return ms.repo.Mint.Update().
		SetStatus(status).
		Where(mintmodel.ID(mintID)).
		Exec(ctx)
}

func (ms *MintService) queryUnprocessedMints(ctx context.Context) ([]*ent.Mint, error) {
	return ms.repo.Mint.Query().
		Where(mintmodel.StatusNotIn(
			mintmodel.StatusSuccess,
			mintmodel.StatusRefunded,
		)).
		WithPegin().
		All(ctx)
}
