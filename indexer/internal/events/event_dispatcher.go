package events

import (
	"context"
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/tontx"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"golang.org/x/sync/errgroup"
)

type EventDispatcher struct {
	inChan      <-chan *ton.RawEvent
	tonTxWriter *tontx.TonTxWriter
	eventWriter *EventWriter
	parsers     map[string]ton.EventParserInterface
}

func NewEventDispatcher(
	inChan <-chan *ton.RawEvent,
	tonTxWriter *tontx.TonTxWriter,
	eventWriter *EventWriter,
	parsers map[string]ton.EventParserInterface,
) *EventDispatcher {
	return &EventDispatcher{
		inChan:      inChan,
		tonTxWriter: tonTxWriter,
		eventWriter: eventWriter,
		parsers:     parsers,
	}
}

func (ed *EventDispatcher) Work(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(64)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case rawEvent, ok := <-ed.inChan:
			if !ok {
				return g.Wait()
			}

			g.Go(func() error {
				return ed.handleEvent(rawEvent)
			})
		}
	}
}

func (ed *EventDispatcher) handleEvent(rawEvent *ton.RawEvent) error {
	parser, exists := ed.parsers[rawEvent.Addr.String()]
	if !exists {
		return ed.formatParserNotFoundError(rawEvent.Addr)
	}

	event, err := parser.Parse(rawEvent)
	if err != nil {
		return err
	}

	tonTx, err := ed.tonTxWriter.Write(event)

	ok, err := ed.handleTonTxWriteError(err)
	if !ok {
		if err != nil {
			return err
		}
		if event.GetEventID() != 0xca444ce6 && event.GetEventID() != 0x27756729 {
			return nil
		}
		tonTx, err := ed.tonTxWriter.GetTonTxWithoutRelationsByHash(fmt.Sprintf("%x", event.GetRaw().TxHash))
		if err != nil {
			return nil
		}
		return ed.eventWriter.Write(tonTx, event)
	}

	return ed.eventWriter.Write(tonTx, event)
}
