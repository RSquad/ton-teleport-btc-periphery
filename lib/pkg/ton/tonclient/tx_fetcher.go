package tonclient

import (
	"context"
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
)

type TxFetcher struct {
	tonClient *TonClient
}

func NewTxFetcher(tonClient *TonClient) *TxFetcher {
	return &TxFetcher{
		tonClient: tonClient,
	}
}

func (tf *TxFetcher) Fetch(ctx context.Context, addr *address.Address, lt uint64, hash []byte, limit int) ([]*tlb.Transaction, error) {
	if limit == 0 {
		limit = 64
	}
	txs, err := tf.tonClient.API.WithRetry(3).ListTransactions(ctx, addr, uint32(limit), lt, hash)
	if err != nil {
		return nil, fmt.Errorf("error fetching txs (addr=%v, txhash=%x, txlt=%v): %w", utils.AddrToRawString(addr), hash, lt, err)
	}
	return txs, nil
}
