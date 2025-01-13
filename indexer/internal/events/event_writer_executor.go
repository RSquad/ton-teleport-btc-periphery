package events

import (
	"log"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
)

type EventWriterExecutor struct {
	eventWriter *EventWriter
}

func NewEventWriterExecutor(eventWriter *EventWriter) *EventWriterExecutor {
	return &EventWriterExecutor{
		eventWriter: eventWriter,
	}
}

func (epe *EventWriterExecutor) Run(eventChan <-chan ton.EventInterface) {
	for event := range eventChan {
		err := epe.eventWriter.Write(event)
		if err != nil {
			log.Printf("error writing event: %v", err)
			continue
		}
	}
}
