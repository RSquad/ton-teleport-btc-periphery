package tontx

import (
	"context"
	"fmt"

	ent "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
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
