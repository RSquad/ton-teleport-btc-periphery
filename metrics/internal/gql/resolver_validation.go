package gql

import (
	"context"
	"fmt"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
	ent "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
	internalkeymodel "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/internalkey"
	peginmodel "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/pegin"
)

func (c *Resolver) findInternalKey(
	ctx context.Context,
	internalKey string,
) (*ent.InternalKey, error) {
	repo := generated.FromContext(ctx)
	return repo.InternalKey.Query().Where(internalkeymodel.KeyEQ(internalKey)).Only(ctx)
}

func (c *Resolver) bitcoinTxExists(
	bitcoinTxIDStr string,
) (bool, *btcjson.TxRawResult, error) {
	bitcoinTxID, err := chainhash.NewHashFromStr(bitcoinTxIDStr)
	if err != nil {
		return false, nil, fmt.Errorf("invalid bitcoin tx id: %w", err)
	}
	bitcoinTx, err := c.bitcoinClient.GetRawTransactionVerbose(bitcoinTxID)
	if err != nil {
		return false, nil, fmt.Errorf("bitcoin tx not found: %w", err)
	}
	return true, bitcoinTx, nil
}

func (c *Resolver) peginExists(
	ctx context.Context,
	bitcoinTxIDStr string,
) (bool, error) {
	repo := generated.FromContext(ctx)
	exists, err := repo.Pegin.Query().Where(peginmodel.BitcoinTxIDEQ(bitcoinTxIDStr)).Exist(ctx)
	return exists, err
}
