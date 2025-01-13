package events

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"

	ent "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/pegin"
	enttontx "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/tontx"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/pegout"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinatorcontract"
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

func (ew *EventWriter) Write(event ton.EventInterface) error {
	exists, err := ew.checkTonTxExists(event)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	tonTx, err := ew.writeTonTx(event)
	if err != nil {
		return err
	}

	err = ew.writeEvent(tonTx, event)
	if err != nil {
		return err
	}

	log.Printf("event wrote (id=%x, txhash=%x)", event.GetEventID(), event.GetRaw().TxHash)

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
	case *coordinatorcontract.DKGCompletedEvent:
		return ew.writeInternalKey(tonTx, event)
	}
	return fmt.Errorf(
		"failed to write event: unknown event type (id=%x, txhash=%x)",
		event.GetEventID(), event.GetRaw().TxHash,
	)
}

func (ew *EventWriter) writeMint(tonTx *ent.TonTx, event *teleportcontract.MintEvent) error {
	return ew.write(func(tx *ent.Tx) error {
		existingPegin, _ := tx.Pegin.Query().
			Where(pegin.BitcoinTxIDEQ(event.BitcoinTxID.String())).
			Only(ew.ctx)

		if existingPegin != nil {
			_, err := tx.Mint.UpdateOne(existingPegin.Edges.Mint).
				SetStatus("SUCCESS").
				SetTonTx(tonTx).
				Save(ew.ctx)
			return err
		}

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
			SetExternalID(int64(event.ID)).
			SetSenderAddr(utils.AddrToRawString(event.SenderAddr)).
			SetAmount(event.Amount.String()).
			SetBitcoinScript(hex.EncodeToString(event.BitcoinScript)).
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
			SetExternalID(int64(event.ID)).
			SetAmount(event.Amount.String()).
			SetBitcoinTxID(event.BitcoinTxID.String()).
			SetBitcoinScript(hex.EncodeToString(event.BitcoinScript)).
			SetTonTx(tonTx).
			SetPegout(pegout).
			Save(ew.ctx)
		return err
	})
}

func (ew *EventWriter) writeInternalKey(tonTx *ent.TonTx, event *coordinatorcontract.DKGCompletedEvent) error {
	_, err := ew.repo.InternalKey.Create().
		SetCompletedAt(event.CompletedAt).
		SetKey(hex.EncodeToString(event.Key)).
		SetTonTx(tonTx).
		Save(ew.ctx)
	return err
}

func (ew *EventWriter) checkTonTxExists(event ton.EventInterface) (bool, error) {
	rawEvent := event.GetRaw()
	exists, err := ew.repo.TonTx.Query().
		Where(enttontx.Hash(fmt.Sprintf("%x", rawEvent.TxHash))).
		Exist(ew.ctx)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (ew *EventWriter) writeTonTx(event ton.EventInterface) (*ent.TonTx, error) {
	rawEvent := event.GetRaw()
	return ew.repo.TonTx.Create().
		SetHash(fmt.Sprintf("%x", rawEvent.TxHash)).
		SetCreatedAt(rawEvent.TxUtime).
		Save(ew.ctx)
}
