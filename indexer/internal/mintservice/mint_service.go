package mintservice

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/bitcoinclientcontract"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	ent "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
	internalkeymodel "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/internalkey"
	mintmodel "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/mint"
	entpegin "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/pegin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/pegincontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	tonclient "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

const (
	defaultLoopInterval = 3 * time.Second
	semaphoreLimit      = 32
)

type MintService struct {
	repo                  *ent.Client
	bitcoinClient         *bitcoin.Client
	tonClient             *tonclient.TonClient
	teleportContract      *teleportcontract.TeleportContract
	bitcoinClientContract *bitcoinclientcontract.BitcoinClientContract
	batchSender           *BatchSender
	indexerWalletAddress  string
	confirmationsNeeded   int64
}

func New(
	repo *ent.Client,
	bitcoinClient *bitcoin.Client,
	tonClient *tonclient.TonClient,
	teleportContract *teleportcontract.TeleportContract,
	bitcoinClientContract *bitcoinclientcontract.BitcoinClientContract,
	batchSender *BatchSender,
	indexerWalletAddress string,
) (*MintService, error) {
	confirmationsNeeded, err := bitcoinClientContract.GetConfirmationsNeeded()
	if err != nil {
		return nil, fmt.Errorf("failed to get number confirmations needed: %w", err)
	}

	return &MintService{
		repo:                  repo,
		bitcoinClient:         bitcoinClient,
		tonClient:             tonClient,
		teleportContract:      teleportContract,
		confirmationsNeeded:   confirmationsNeeded,
		bitcoinClientContract: bitcoinClientContract,
		indexerWalletAddress:  indexerWalletAddress,
		batchSender:           batchSender,
	}, nil
}

func (ms *MintService) Work(ctx context.Context) error {
	ms.logStartWork()

	var wg sync.WaitGroup
	wg.Add(2)

	go ms.continuouslyProcessPendingMints(ctx, &wg)
	go ms.continuouslyProcessRefundMints(ctx, &wg)

	<-ctx.Done()
	logContextCancelled()

	wg.Wait()
	ms.logFinishWork(ctx.Err())

	return ctx.Err()
}

func (ms *MintService) continuouslyProcessPendingMints(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ms.logStartPendingWork()

	var err error
	defer func() { ms.logFinishPendingWork(err) }()

	ticker := time.NewTicker(defaultLoopInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return
		case <-ticker.C:
			cycleErr := ms.executePendingMintsCycle(ctx)
			if cycleErr != nil {
				logPendingCycleError(cycleErr)
			}
		}
	}
}

func (ms *MintService) executePendingMintsCycle(ctx context.Context) (err error) {
	start := time.Now()
	var processedCount int
	defer func() {
		logFinishProcessingPendingMints(time.Since(start), err, processedCount)
	}()
	logStartProcessingPendingMints()

	mints, err := ms.repo.Mint.Query().
		Where(mintmodel.StatusEQ(mintmodel.StatusPending)).
		WithPegin().
		All(ctx)
	if err != nil {
		return fmt.Errorf(errQueryPendingMints, err)
	}

	if len(mints) == 0 {
		logNoPendingMints()
		return nil
	}
	logPendingMintsReceived(len(mints))
	processedCount = len(mints)

	block, err := ms.tonClient.API.CurrentMasterchainInfo(ctx)
	if err != nil {
		return fmt.Errorf(tonclient.ErrGetCurrentMasterchainInfo, err)
	}

	teleportStorage, err := ms.teleportContract.GetStorage(block)
	if err != nil {
		return fmt.Errorf(teleportcontract.ErrGetStorage, err)
	}

	lastConfirmedBlockHash, err := ms.bitcoinClientContract.GetLastConfirmedBlockHash()
	if err != nil {
		return fmt.Errorf("failed to get last confirmed block hash: %w", err)
	}

	lastConfirmedBlockHeight, err := ms.bitcoinClient.GetBlockHeightByHash(lastConfirmedBlockHash)
	if err != nil {
		return fmt.Errorf("failed to get last confirmed block height: %w", err)
	}

	latestInternalKey, err := ms.repo.InternalKey.Query().
		Order(ent.Desc(internalkeymodel.FieldCompletedAt)).
		First(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			return fmt.Errorf(errQueryLatestInternalKey, err)
		}
		logNoInternalKeysFoundWarning()
		latestInternalKey = nil
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, semaphoreLimit)

	for i, mint := range mints {
		wg.Add(1)
		sem <- struct{}{}
		go func(m *ent.Mint, idx int) {
			defer wg.Done()
			defer func() { <-sem }()
			pending, err := ms.handlePendingMint(ctx, m, teleportStorage.PeginContractCode, block, latestInternalKey)
			if err != nil {
				logFailedProcessPendingMint(err, m.ID)
			}
			if pending {
				err = ms.handleUnconfirmedMint(ctx, m, lastConfirmedBlockHeight)
				if err != nil {
					logFailedProcessUnconfirmedMint(err, m.ID)
				}
			}

			logPendingMintsProcessingProgress(idx+1, len(mints))
		}(mint, i)
	}

	wg.Wait()

	sendedMessagesData, err := ms.batchSender.Send()
	if err != nil {
		return fmt.Errorf("failed to send messages: %w", err)
	}
	logSendedMessages(len(sendedMessagesData))

	for i, message := range sendedMessagesData {
		wg.Add(1)
		go func(mes *MessageWithTxHash, idx int) {
			defer wg.Done()
			ms.waitForDepositAndActivation(ctx, message.Mint, teleportStorage.PeginContractCode, message.TxHash, latestInternalKey)
		}(message, i)
	}
	wg.Wait()

	return nil
}

func (ms *MintService) continuouslyProcessRefundMints(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ms.logStartRefundWork()
	var err error
	defer func() { ms.logFinishRefundWork(err) }()

	ticker := time.NewTicker(defaultLoopInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return
		case <-ticker.C:
			cycleErr := ms.executeRefundMintsCycle(ctx)
			if cycleErr != nil {
				logRefundCycleError(cycleErr)
			}
		}
	}
}

func (ms *MintService) executeRefundMintsCycle(ctx context.Context) (err error) {
	start := time.Now()
	var processedCount int
	defer func() {
		logFinishProcessingRefundMints(time.Since(start), err, processedCount)
	}()
	logStartProcessingRefundMints()

	mints, err := ms.repo.Mint.Query().
		Where(mintmodel.StatusEQ(mintmodel.StatusRefund)).
		WithPegin().
		All(ctx)
	if err != nil {
		return fmt.Errorf(errQueryRefundMints, err)
	}

	if len(mints) == 0 {
		logNoRefundMints()
		return nil
	}
	logRefundMintsReceived(len(mints))
	processedCount = len(mints)

	var wg sync.WaitGroup
	sem := make(chan struct{}, semaphoreLimit)

	for i, mint := range mints {
		wg.Add(1)
		sem <- struct{}{}
		go func(m *ent.Mint, idx int) {
			defer wg.Done()
			defer func() { <-sem }()
			err := ms.handleRefundMint(ctx, m)
			if err != nil {
				logFailedProcessRefundMint(err, m.ID)
			}
			logRefundMintsProcessingProgress(idx+1, len(mints))
		}(mint, i)
	}

	wg.Wait()
	return nil
}

func (ms *MintService) isPeginContractActive(ctx context.Context, peginContractCode *cell.Cell, block *ton.BlockIDExt, bitcoinTxID *chainhash.Hash) (bool, error) {
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
		return false, fmt.Errorf(errCreatePeginContractFromStateInit, err)
	}

	peginState, err := ms.tonClient.API.GetAccount(ctx, block, peginContract.Addr)
	if err != nil {
		return false, fmt.Errorf(errGetPeginContractAccountState, err)
	}

	if peginState.IsActive {
		return true, nil
	}

	return false, nil
}

func (ms *MintService) handleUnconfirmedMint(
	ctx context.Context,
	mint *ent.Mint,
	lastConfirmedBlockHeight int64,
) error {
	bitcoinTxID, err := chainhash.NewHashFromStr(mint.Edges.Pegin.BitcoinTxID)
	if err != nil {
		return fmt.Errorf(errCalcBitcoinTxID, err)
	}

	confirmationsAchieved, bitcoinTXBlockHash, err := ms.areConfirmationsAchieved(bitcoinTxID, lastConfirmedBlockHeight)
	if err != nil {
		return fmt.Errorf("failed to wait for confirmations: %w", err)
	}

	if !confirmationsAchieved {
		return nil
	}

	receiverAddress, recoveryKey, txHex, txProof, err := ms.fetchTransactionInfo(ctx, bitcoinTxID, bitcoinTXBlockHash)
	if err != nil {
		return fmt.Errorf("failed to fetch transaction info: %w", err)
	}

	message, err := ms.teleportContract.BuildSendDepositMessage(ctx, bitcoinTXBlockHash, receiverAddress, ms.indexerWalletAddress, recoveryKey, txHex, txProof)
	if err != nil {
		return fmt.Errorf("failed to send deposit: %w", err)
	}

	messageData := &MessageWithTxHash{
		Mint:    mint,
		Message: message,
		TxHash:  bitcoinTXBlockHash,
	}

	ms.batchSender.messagesForSending <- messageData

	return nil
}

func (ms *MintService) handlePendingMint(
	ctx context.Context,
	mint *ent.Mint,
	peginContractCode *cell.Cell,
	block *ton.BlockIDExt,
	latestInternalKey *ent.InternalKey,
) (bool, error) {
	bitcoinTxID, err := chainhash.NewHashFromStr(mint.Edges.Pegin.BitcoinTxID)
	if err != nil {
		return false, fmt.Errorf(errCalcBitcoinTxID, err)
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
		return false, fmt.Errorf(errCreatePeginContractFromStateInit, err)
	}

	peginState, err := ms.tonClient.API.GetAccount(ctx, block, peginContract.Addr)
	if err != nil {
		return false, fmt.Errorf(errGetPeginContractAccountState, err)
	}

	if peginState.IsActive {
		return false, ms.updateMintStatus(ctx, mint.ID, mintmodel.StatusSuccess)
	}

	if latestInternalKey == nil {
		return false, ms.updateMintStatus(ctx, mint.ID, mintmodel.StatusRefund)
	}

	internalKey, err := ms.repo.InternalKey.Query().
		Where(internalkeymodel.KeyEQ(mint.Edges.Pegin.InternalKey)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			logPendingMintPeginInternalKeyNotFound(err, mint.ID, mint.Edges.Pegin.InternalKey)
			return false, ms.updateMintStatus(ctx, mint.ID, mintmodel.StatusRefund)
		}
		return false, fmt.Errorf(errQueryInternalKey, err)
	}

	if !internalKey.CompletedAt.Equal(latestInternalKey.CompletedAt) {
		return false, ms.updateMintStatus(ctx, mint.ID, mintmodel.StatusRefund)
	}

	return true, nil
}

func (ms *MintService) handleRefundMint(ctx context.Context, mint *ent.Mint) error {
	bitcoinTxID, err := chainhash.NewHashFromStr(mint.Edges.Pegin.BitcoinTxID)
	if err != nil {
		logRefundMintFailedParseBitcoinTxID(err, mint.ID, mint.Edges.Pegin.BitcoinTxID)
		return err
	}

	out, err := ms.bitcoinClient.GetTxOut(bitcoinTxID, uint32(mint.Edges.Pegin.VoutIndex), true)
	if err != nil {
		logRefundMintFailedGetTxOut(err, mint.ID)
		return err
	}

	if out == nil {
		return ms.updateMintStatus(ctx, mint.ID, mintmodel.StatusRefunded)
	}

	return nil
}

func (ms *MintService) updateMintStatus(ctx context.Context, mintID int, status mintmodel.Status) error {
	err := ms.repo.Mint.Update().
		SetStatus(status).
		Where(mintmodel.ID(mintID)).
		Exec(ctx)
	if err != nil {
		logFailedUpdateMintStatus(err, mintID, status)
		return err
	}
	logMintStatusUpdated(mintID, status)
	return nil
}

func (ms *MintService) areConfirmationsAchieved(bitcoinTxID *chainhash.Hash, lastConfirmedBlockHeight int64) (bool, *chainhash.Hash, error) {
	bitcoinTXBlockHash, err := ms.bitcoinClient.GetBlockHashByTxID(bitcoinTxID)
	if err != nil {
		if strings.Contains(err.Error(), "not yet mined or blockhash not available") {
			return false, nil, nil
		}
		return false, nil, fmt.Errorf("failed to get block hash by tx ID: %w", err)
	}

	bitcoinTXBlockHeight, err := ms.bitcoinClient.GetBlockHeightByHash(bitcoinTXBlockHash)
	if err != nil {
		return false, nil, fmt.Errorf("failed to get block height: %w", err)
	}

	if bitcoinTXBlockHeight <= lastConfirmedBlockHeight {
		return true, bitcoinTXBlockHash, nil
	}

	return false, nil, nil
}

func (ms *MintService) waitForDepositAndActivation(
	ctx context.Context,
	mint *ent.Mint,
	peginContractCode *cell.Cell,
	bitcoinTxID *chainhash.Hash,
	latestInternalKey *ent.InternalKey,
) {
	for i := 0; i < 10; i++ {
		select {
		case <-ctx.Done():
			ms.logDepositContextCancelled(mint.ID, ctx.Err())
			return
		default:
		}

		block, err := ms.tonClient.API.CurrentMasterchainInfo(ctx)
		if err != nil {
			ms.logPeginActivationCheckFailed(mint.ID, err)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		isPeginContractActive, err := ms.isPeginContractActive(ctx, peginContractCode, block, bitcoinTxID)
		if err != nil {
			ms.logPeginActivationCheckFailed(mint.ID, err)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		if isPeginContractActive {
			ms.handlePendingMint(ctx, mint, peginContractCode, block, latestInternalKey)
			return
		}

		time.Sleep(time.Duration(i+1) * time.Second)
	}

	ms.logPeginActivationTimeout(mint.ID)
}

func (ms *MintService) fetchTransactionInfo(
	ctx context.Context,
	bitcoinTxID *chainhash.Hash,
	blockHash *chainhash.Hash,
) (
	receiverAddress string,
	recoveryKey string,
	txHex string,
	txProof []byte,
	err error,
) {
	txProofStr, err := ms.bitcoinClient.GetTxProof(bitcoinTxID, blockHash)
	if err != nil {
		return "", "", "", nil, fmt.Errorf("failed to get tx proof: %w", err)
	}

	txProof, err = hex.DecodeString(txProofStr)
	if err != nil {
		return "", "", "", nil, fmt.Errorf("failed to decode tx proof: %w", err)
	}

	txVerbose, err := ms.bitcoinClient.GetRawTransactionVerbose(bitcoinTxID)
	if err != nil {
		return "", "", "", nil, fmt.Errorf("failed to get tx verbose: %w", err)
	}
	txHex = txVerbose.Hex

	existingPegin, err := ms.repo.Pegin.
		Query().
		Where(entpegin.BitcoinTxIDEQ(bitcoinTxID.String())).
		Only(ctx)
	if err != nil {
		return "", "", "", nil, fmt.Errorf("failed to get pegin: %w", err)
	}

	return existingPegin.ReceiverAddr, existingPegin.RecoveryKey, txHex, txProof, nil
}
