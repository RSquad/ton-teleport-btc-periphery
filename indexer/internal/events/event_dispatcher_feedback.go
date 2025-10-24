package events

import (
	"fmt"
	"strings"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
)

func (ed *EventDispatcher) handleTonTxWriteError(err error) (bool, error) {
	if err == nil {
		return true, nil
	}
	if strings.Contains(err.Error(), "duplicate key value violates unique constraint \"ton_txes_hash_key\"") {
		logDuplicateKeyError(err)
		return false, nil
	}
	logTonTxWriteError(err)
	return false, err
}

func (ed *EventDispatcher) formatParserNotFoundError(addr *address.Address) error {
	return fmt.Errorf("no parser found for address %s", utils.AddrToRawString(addr))
}

func logDuplicateKeyError(err error) {
	logger.Log.Error().
		Str("component", "EventDispatcher").
		Err(err).
		Msg("Duplicate key error")
}

func logTonTxWriteError(err error) {
	logger.Log.Error().
		Str("component", "EventDispatcher").
		Err(err).
		Msg("Failed to write ton tx")
}

func logParserNotFoundError(addr *address.Address) {
	logger.Log.Error().
		Str("component", "EventDispatcher").
		Str("event_addr", utils.AddrToRawString(addr)).
		Msg("Parser not found")
}

func logUnknownEventTypeError(event *ton.RawEvent) {
	logger.Log.Error().
		Str("component", "EventDispatcher").
		Str("event_addr", utils.AddrToRawString(event.Addr)).
		Str("event_tx_hash", string(event.TxHash)).
		Msg("Unknown event type")
}
