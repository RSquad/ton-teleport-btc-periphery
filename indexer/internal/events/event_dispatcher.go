package events

import (
	"context"
	"strings"

	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/tontx"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
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

func (ed *EventDispatcher) Work(ctx context.Context) (err error) {
	logger.Log.Info().Str("component", "EventDispatcher").Msg("started")
	defer func() {
		logger.Log.Info().Str("component", "EventDispatcher").Err(err).Msg("finished")
	}()
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(64)

	for {
		select {
		case <-ctx.Done():
			g.Wait()
			return ctx.Err()
		case rawEvent, ok := <-ed.inChan:
			// if channel is closed, wait for all goroutines to finish
			if !ok {
				logger.Log.Debug().Str("component", "EventDispatcher").Msg("event channel closed")
				return g.Wait()
			}

			g.Go(func() error {
				if err := ed.handleEvent(rawEvent); err != nil {
					logger.Log.Error().Str("component", "EventDispatcher").Err(err).Msg("event handler failed")
				}
				return nil
			})
		}
	}
}

func (ed *EventDispatcher) handleEvent(rawEvent *ton.RawEvent) error {
	parser, exists := ed.parsers[rawEvent.Addr.String()]
	if !exists {
		logParserNotFoundError(rawEvent.Addr)
		return ed.formatParserNotFoundError(rawEvent.Addr)
	}

	tonTx, err := ed.tonTxWriter.Write(rawEvent)
	if err != nil {
		return err
	}

	event, err := parser.Parse(rawEvent)
	if err != nil {
		if strings.Contains(err.Error(), "unknown event type") {
			logUnknownEventTypeError(rawEvent)
			return nil
		}

		return err
	}

	return ed.eventWriter.Write(tonTx, event)
}
