package tontx

import (
	"context"
	"fmt"

	ent "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/tontx"
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
	event ton.EventInterface,
) (*ent.TonTx, error) {
	rawEvent := event.GetRaw()
	return ew.repo.TonTx.Create().
		SetHash(fmt.Sprintf("%x", rawEvent.TxHash)).
		SetCreatedAt(rawEvent.TxUtime).
		Save(ew.ctx)
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
