package ton

import (
	"log"
)

type EventParserExecutor struct {
	eventParser EventParserInterface
}

func NewEventParserExecutor(eventParser EventParserInterface) *EventParserExecutor {
	return &EventParserExecutor{
		eventParser: eventParser,
	}
}

func (lpe *EventParserExecutor) Run(rawEventsChan <-chan *RawEvent, parsedEventChan chan<- EventInterface) {
	for rawEvent := range rawEventsChan {
		parsedEvent, err := lpe.eventParser.Parse(rawEvent)
		if err != nil {
			log.Printf("error parsing event: %v", err)
			continue
		}
		parsedEventChan <- parsedEvent
	}
}
