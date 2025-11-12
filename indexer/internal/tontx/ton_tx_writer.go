package tontx

import (
	"context"
	"fmt"

	ent "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/tontx"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
)

type TonTxWriter struct {
	ctx  context.Context
	repo *ent.Client
}

func NewTonTxWriter(
	ctx context.Context,
	repo *ent.Client,
) *TonTxWriter {
	return &TonTxWriter{
		ctx:  ctx,
		repo: repo,
	}
}

func (ew *TonTxWriter) Write(
	rawEvent *ton.RawEvent,
) (*ent.TonTx, error) {
	hash := fmt.Sprintf("%x", rawEvent.TxHash)
	logger.Log.Debug().Str("component", "TonTxWriter").Str("tx_hash", hash).Msg("query")
	tonTx, err := ew.repo.TonTx.Query().Where(tontx.Hash(hash)).First(ew.ctx)
	if ent.IsNotFound(err) {
		return ew.repo.TonTx.Create().
			SetHash(hash).
			SetCreatedAt(rawEvent.TxUtime).
			Save(ew.ctx)
	} else {
		logger.Log.Warn().Str("component", "TonTxWriter").Str("tx_hash", hash).Msg("already exists")
	}
	return tonTx, err
}

func (ew *TonTxWriter) GetTonTxWithoutRelationsByHash(hash string) (*ent.TonTx, error) {
	return ew.repo.TonTx.
		Query().
		Where(tontx.Hash(hash),
			tontx.Not(tontx.HasBurn()),
			tontx.Not(tontx.HasReinit()),
			tontx.Not(tontx.HasMint()),
			tontx.Not(tontx.HasInternalKey()),
		).First(ew.ctx)
}
