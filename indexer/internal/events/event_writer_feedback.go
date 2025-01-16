package events

import (
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
)

func (ew *EventWriter) logEventWritten(event ton.EventInterface) {
	logger.Log.Info().
		Str("component", "EventWriter").
		Str("event_id", fmt.Sprintf("%x", event.GetEventID())).
		Str("tx_hash", fmt.Sprintf("%x", event.GetRaw().TxHash)).
		Msg("Event written")
}

func (ew *EventWriter) formatUnknownEventError(event ton.EventInterface) error {
	return fmt.Errorf(
		"failed to write event: unknown event with id %x in tx with hash=%x",
		event.GetEventID(), event.GetRaw().TxHash,
	)
}
