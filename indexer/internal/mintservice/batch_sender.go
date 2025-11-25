package mintservice

import (
	"context"
	"fmt"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	tonclient "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

type BatchSender struct {
	tonClient          *tonclient.TonClient
	highLoadWallet     *wallet.Wallet
	messagesForSending chan Message
}

func NewBatchSender(highLoadWallet *wallet.Wallet, tonClient *tonclient.TonClient) *BatchSender {
	batchSender := &BatchSender{
		highLoadWallet:     highLoadWallet,
		tonClient:          tonClient,
		messagesForSending: make(chan Message, 32),
	}
	return batchSender
}

type Message interface {
	GetMessage() *wallet.Message
}

// max messages size per transaction for highload wallet is 64KB
const maxBatchSize = 64 * 1024

func (b *BatchSender) Send() ([]Message, error) {
	capacityCount, err := b.GetMessageCapacity(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get message capacity: %w", err)
	}

	if len(b.messagesForSending) == 0 {
		return nil, nil
	}

	capacitySize := maxBatchSize
	offset := 0
	sentMessages := make([]Message, 0)
	batch := make([]*wallet.Message, 0)

	for message := range b.messagesForSending {
		msg := message.GetMessage()
		size, err := b.messageSize(msg)
		if err != nil {
			logger.Log.Error().Err(err).
				Str("component", "BatchSender").
				Msg("failed to get message size for message")
			continue
		}
		capacityCount -= 1
		capacitySize -= size
		if capacitySize <= 0 || capacityCount <= 0 {
			b.highLoadWallet.SendManyWaitTxHash(context.Background(), batch[offset:])
			time.Sleep(1000 * time.Millisecond)
			offset = len(batch)
			capacitySize = maxBatchSize
		} else {
			batch = append(batch, msg)
			sentMessages = append(sentMessages, message)
		}

		if len(b.messagesForSending) == 0 {
			break
		}
	}

	if len(batch[offset:]) > 0 {
		b.highLoadWallet.SendManyWaitTxHash(context.Background(), batch[offset:])
	}
	return sentMessages, nil
}

func (b *BatchSender) EnqueueMessage(message Message) {
	b.messagesForSending <- message
}

func (b *BatchSender) messageSize(message *wallet.Message) (int, error) {
	cell, err := tlb.ToCell(message.InternalMessage)
	if err != nil {
		return 0, err
	}

	// +1 is a send_mode byte
	return len(cell.ToBOC()) + 1, nil
}

func (b *BatchSender) GetMessageCapacity(ctx context.Context) (int, error) {
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

	// 1 Ton is reserved for gas and storage
	messageCapacity := int((balanceNano.Int64() - TONInNano) / TONInNano)

	if messageCapacity < 10 {
		logger.Log.Warn().
			Str("component", "BatchSender").
			Int("message_capacity", messageCapacity).
			Msg("wallet message capacity is too low")
	}

	return messageCapacity, nil
}
