package events

import (
	"context"
	"encoding/hex"
	"log"

	ent "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/pegin"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/pegout"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
)

type EventWriter struct {
	ctx          context.Context
	repo         *ent.Client
	pegoutWriter *pegout.PegoutWriter
}

func NewEventWriter(
	ctx context.Context,
	repo *ent.Client,
	pegoutWriter *pegout.PegoutWriter,
) *EventWriter {
	return &EventWriter{
		ctx, repo, pegoutWriter,
	}
}

func (ew *EventWriter) write(fn func(tx *ent.Tx) error) error {
	tx, err := ew.repo.Tx(ew.ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()
	return fn(tx)
}

func (ew *EventWriter) Write(tonTx *ent.TonTx, event ton.EventInterface) error {
	err := ew.writeEvent(tonTx, event)
	if err != nil {
		return err
	}

	ew.logEventWritten(event)

	return nil
}

func (ew *EventWriter) writeEvent(tonTx *ent.TonTx, event ton.EventInterface) error {
	switch event := event.(type) {
	case *teleportcontract.MintEvent:
		return ew.writeMint(tonTx, event)
	case *teleportcontract.BurnEvent:
		return ew.writeBurn(tonTx, event)
	case *teleportcontract.ReinitEvent:
		return ew.writeReinit(tonTx, event)
	case *coordinator.DKGCompletedEvent:
		return ew.writeInternalKey(tonTx, event)
	}
	return ew.formatUnknownEventError(event)
}

func (ew *EventWriter) writeMint(tonTx *ent.TonTx, event *teleportcontract.MintEvent) error {
	existingPegin, _ := ew.repo.Pegin.Query().
		Where(pegin.BitcoinTxIDEQ(event.BitcoinTxID.String())).
		WithMint().
		Only(ew.ctx)

	if existingPegin != nil {
		log.Printf(
			"updating mint status to SUCCESS (mintid=%d, txhash=%x)",
			existingPegin.Edges.Mint.ID, event.GetRaw().TxHash,
		)
		_, err := ew.repo.Mint.UpdateOne(existingPegin.Edges.Mint).
			SetStatus("SUCCESS").
			SetTonTx(tonTx).
			Save(ew.ctx)
		return err
	}

	return ew.write(func(tx *ent.Tx) error {
		mint, err := tx.Mint.Create().
			SetAmount(event.Amount.String()).
			SetStatus("SUCCESS").
			SetCreatedAt(tonTx.CreatedAt).
			SetTonTx(tonTx).
			Save(ew.ctx)
		if err != nil {
			return err
		}

		_, err = tx.Pegin.Create().
			SetReceiverAddr(utils.AddrToRawString(event.ReceiverAddr)).
			SetBitcoinTxID(event.BitcoinTxID.String()).
			SetMint(mint).
			Save(ew.ctx)
		return err
	})
}

func (ew *EventWriter) writeBurn(tonTx *ent.TonTx, event *teleportcontract.BurnEvent) error {
	return ew.write(func(tx *ent.Tx) error {
		pegout, err := ew.pegoutWriter.WriteFromEvent(event)
		if err != nil {
			return err
		}
		_, err = tx.Burn.Create().
			SetSenderAddr(utils.AddrToRawString(event.SenderAddr)).
			SetAmount(event.Amount.String()).
			SetTonTx(tonTx).
			SetPegout(pegout).
			Save(ew.ctx)
		return err
	})
}

func (ew *EventWriter) writeReinit(tonTx *ent.TonTx, event *teleportcontract.ReinitEvent) error {
	return ew.write(func(tx *ent.Tx) error {
		pegout, err := ew.pegoutWriter.WriteFromEvent(event)
		if err != nil {
			return err
		}
		_, err = tx.Reinit.Create().
			SetAmount(event.Amount.String()).
			SetNewInternalKey(hex.EncodeToString(event.NewInternalKey)).
			SetTonTx(tonTx).
			SetPegout(pegout).
			Save(ew.ctx)
		return err
	})
}

func (ew *EventWriter) writeInternalKey(tonTx *ent.TonTx, event *coordinator.DKGCompletedEvent) error {
	_, err := ew.repo.InternalKey.Create().
		SetCompletedAt(event.CompletedAt).
		SetKey(hex.EncodeToString(event.Key)).
		SetTonTx(tonTx).
		Save(ew.ctx)
	return err
}
