package mintservice

import (
	"context"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	ent "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
	tonclient "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

type BatchSender struct {
	tonClient          *tonclient.TonClient
	highLoadWallet     *wallet.Wallet
	messagesForSending chan *MessageWithTxHash
}

func NewBatchSender(highLoadWallet *wallet.Wallet, tonClient *tonclient.TonClient) *BatchSender {
	batchSender := &BatchSender{
		highLoadWallet:     highLoadWallet,
		tonClient:          tonClient,
		messagesForSending: make(chan *MessageWithTxHash, 100),
	}
	return batchSender
}

type MessageWithTxHash struct {
	Mint    *ent.Mint
	Message *wallet.Message
	TxHash  *chainhash.Hash
}

func (b *BatchSender) Send() ([]*MessageWithTxHash, error) {
	batchSize := 0
	batch := make([]*wallet.Message, 0)
	sendedMessages := make([]*MessageWithTxHash, 0)
	for message := range b.messagesForSending {
		size, err := b.messageSize(message.Message)
		if err != nil {
			return nil, fmt.Errorf("failed to get message size: %w", err)
		}
		if batchSize+size > 64*1024 {
			possibleMessagesNumber, err := b.CheckBalance(context.Background(), len(batch))
			if err != nil {
				if possibleMessagesNumber > 0 {
					b.highLoadWallet.SendManyWaitTxHash(context.Background(), batch[:possibleMessagesNumber])
					return sendedMessages, err
				}
			}
			b.highLoadWallet.SendManyWaitTxHash(context.Background(), batch)
			time.Sleep(1000 * time.Millisecond)
			batch = make([]*wallet.Message, 0)
			batchSize = 0
			continue
		}
		batch = append(batch, message.Message)
		sendedMessages = append(sendedMessages, message)
		if len(b.messagesForSending) == 0 {
			break
		}
	}
	possibleMessagesNumber, err := b.CheckBalance(context.Background(), len(batch))
	if err != nil {
		if possibleMessagesNumber > 0 {
			b.highLoadWallet.SendManyWaitTxHash(context.Background(), batch[:possibleMessagesNumber])
			return sendedMessages, err
		}
		return nil, fmt.Errorf("failed to check balance: %w", err)
	}
	b.highLoadWallet.SendManyWaitTxHash(context.Background(), batch)

	return sendedMessages, nil
}

func (b *BatchSender) messageSize(message *wallet.Message) (int, error) {
	cell, err := tlb.ToCell(message.InternalMessage)
	if err != nil {
		return 0, err
	}

	return len(cell.ToBOC()) + 1, nil
}

func (b *BatchSender) CheckBalance(ctx context.Context, messagesNumber int) (int, error) {
	block, err := b.tonClient.API.CurrentMasterchainInfo(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get current masterchain info: %w", err)
	}

	acc, err := b.tonClient.API.GetAccount(ctx, block, b.highLoadWallet.Address())
	if err != nil {
		return 0, fmt.Errorf("GetAccount: %w", err)
	}

	balanceNano := acc.State.Balance.Nano()

	TONInNano := int64(1_000_000_000)

	if balanceNano.Int64() < TONInNano*int64(messagesNumber) {
		return int(balanceNano.Int64() / TONInNano), fmt.Errorf("not enough balance: %d < %d", balanceNano.Int64(), TONInNano*int64(messagesNumber))
	}

	return messagesNumber, nil
}
