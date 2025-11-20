package teleportcontract

import (
	"context"
	"fmt"
	"sync"
	"time"

	tonclient "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

type EnqueuerMessages struct {
	TonClient       *tonclient.TonClient
	highLoadWallet  *wallet.Wallet
	batchMu         sync.Mutex
	batchMessages   []messageToSend
	batchSize       int
	flushStateMu    sync.Mutex
	isFlushBatching bool
}

func NewEnqueuerMessages(highLoadWallet *wallet.Wallet, tonClient *tonclient.TonClient) *EnqueuerMessages {
	enqueuer := &EnqueuerMessages{
		highLoadWallet: highLoadWallet,
		TonClient:      tonClient,
		batchMu:        sync.Mutex{},
		batchMessages:  []messageToSend{},
		batchSize:      0,
	}
	go enqueuer.StartBatchFlusher()
	return enqueuer
}

func (e *EnqueuerMessages) StartBatchFlusher() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if e.isManualFlushInProgress() {
				continue
			}
			e.flushBatch()
		}
	}()
}

type messageToSend struct {
	Msg       *wallet.Message
	ReplyChan chan SendResult
	Size      int
}

type SendResult struct {
	TxHash []byte
	Err    error
}

func (e *EnqueuerMessages) EnqueueMessage(msg *wallet.Message, size int) (<-chan SendResult, error) {
	replyChan := make(chan SendResult, 1)

	e.batchMu.Lock()

	if len(e.batchMessages) >= 254 || e.batchSize+size > 64*1024 {
		e.batchMu.Unlock()
		e.setManualFlushInProgress(true)
		e.flushBatch()
		e.setManualFlushInProgress(false)
		e.batchMu.Lock()
	}

	e.batchMessages = append(e.batchMessages, messageToSend{
		Msg:       msg,
		ReplyChan: replyChan,
		Size:      size,
	})
	e.batchSize += size

	e.batchMu.Unlock()

	return replyChan, nil
}

func (e *EnqueuerMessages) flushBatch() {
	e.batchMu.Lock()

	if len(e.batchMessages) == 0 {
		e.batchMu.Unlock()
		return
	}

	msgs := make([]*wallet.Message, len(e.batchMessages))
	replyChans := make([]chan SendResult, len(e.batchMessages))
	for i, m := range e.batchMessages {
		msgs[i] = m.Msg
		replyChans[i] = m.ReplyChan
	}

	messagesNumber := len(e.batchMessages)
	fmt.Println("messagesNumber", messagesNumber)
	e.batchMessages = nil
	e.batchSize = 0

	e.batchMu.Unlock()

	if err := e.CheckBalance(context.Background(), messagesNumber); err != nil {
		for _, ch := range replyChans {
			ch <- SendResult{TxHash: nil, Err: err}
			close(ch)
		}
		return
	}

	txHash, err := e.highLoadWallet.SendManyWaitTxHash(context.Background(), msgs)

	for _, ch := range replyChans {
		ch <- SendResult{TxHash: txHash, Err: err}
		close(ch)
	}

}

func (e *EnqueuerMessages) setManualFlushInProgress(val bool) {
	e.flushStateMu.Lock()
	e.isFlushBatching = val
	e.flushStateMu.Unlock()
}

func (e *EnqueuerMessages) isManualFlushInProgress() bool {
	e.flushStateMu.Lock()
	defer e.flushStateMu.Unlock()
	return e.isFlushBatching
}

func (e *EnqueuerMessages) CheckBalance(ctx context.Context, messagesNumber int) error {
	block, err := e.TonClient.API.CurrentMasterchainInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current masterchain info: %w", err)
	}

	acc, err := e.TonClient.API.GetAccount(ctx, block, e.highLoadWallet.Address())
	if err != nil {
		return fmt.Errorf("GetAccount: %w", err)
	}

	balanceNano := acc.State.Balance.Nano()

	TONInNano := int64(1_000_000_000)

	if balanceNano.Int64() < TONInNano*int64(messagesNumber) {
		return fmt.Errorf("not enough balance: %d < %d", balanceNano.Int64(), TONInNano*int64(messagesNumber))
	}

	return nil
}
