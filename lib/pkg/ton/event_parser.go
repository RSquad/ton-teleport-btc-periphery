package ton

type EventParserInterface interface {
	Parse(rawEvent *RawEvent) (EventInterface, error)
}
