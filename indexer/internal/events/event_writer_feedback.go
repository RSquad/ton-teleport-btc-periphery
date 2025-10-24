package events

import (
	"encoding/hex"
	"fmt"

	ent "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
)

func (ew *EventWriter) logEventWritten(event ton.EventInterface) {
	logger.Log.Info().
		Str("component", "EventWriter").
		Str("event_id", fmt.Sprintf("%x", event.GetEventID())).
		Str("tx_hash", hex.EncodeToString(event.GetRaw().TxHash)).
		Msg("Event written")
}

func (ew EventWriter) logMintEventUpdateStatus(mint *ent.Mint, event ton.EventInterface) {
	logger.Log.Info().
		Str("component", "EventWriter").
		Int("mint_id", mint.ID).
		Str("tx_hash", hex.EncodeToString(event.GetRaw().TxHash)).
		Msg("Mint status updated to SUCCESS")
}

func logEventWriteError(err error, event ton.EventInterface) {
	logger.Log.Error().
		Str("component", "EventWriter").
		Err(err).
		Int("event_id", int(event.GetEventID())).
		Msg("Event write error")
}

func logMintUpdateError(err error, event ton.EventInterface, mint *ent.Mint) {
	logger.Log.Error().
		Str("component", "EventWriter").
		Err(err).
		Int("mint_id", mint.ID).
		Str("tx_hash", hex.EncodeToString(event.GetRaw().TxHash)).
		Msg("Mint update write error")
}

func logMintCreateError(err error, event ton.EventInterface, tx *ent.TonTx) {
	logger.Log.Error().
		Str("component", "EventWriter").
		Err(err).
		Int("tx_id", tx.ID).
		Str("tx_hash", hex.EncodeToString(event.GetRaw().TxHash)).
		Msg("Mint create error")
}

func logPeginCreateError(err error, event ton.EventInterface, mint *ent.Mint) {
	logger.Log.Error().
		Str("component", "EventWriter").
		Err(err).
		Int("mint_id", mint.ID).
		Str("tx_hash", hex.EncodeToString(event.GetRaw().TxHash)).
		Msg("Pegin create error")
}

func logBurnCreateError(err error, event ton.EventInterface, pegout *ent.Pegout) {
	logger.Log.Error().
		Str("component", "EventWriter").
		Err(err).
		Int("pegout_id", pegout.ID).
		Str("tx_hash", hex.EncodeToString(event.GetRaw().TxHash)).
		Msg("Burn create error")
}

func logReinitCreateError(err error, event ton.EventInterface, reinit *ent.Reinit) {
	logger.Log.Error().
		Str("component", "EventWriter").
		Err(err).
		Int("reinit_id", reinit.ID).
		Str("tx_hash", hex.EncodeToString(event.GetRaw().TxHash)).
		Msg("Reinit create error")
}

func logInternalKeyCreateError(err error, event *coordinator.DKGCompletedEvent) {
	logger.Log.Error().
		Str("component", "EventWriter").
		Err(err).
		Str("key", hex.EncodeToString(event.Key)).
		Str("tx_hash", hex.EncodeToString(event.GetRaw().TxHash)).
		Msg("InternalKey create error")
}
