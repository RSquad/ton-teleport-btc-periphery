package fetchers

type MetricsPayloadTypeDB int

const (
	PayloadTypeDKG                       MetricsPayloadTypeDB = iota
	PayloadTypePrevDKG                   MetricsPayloadTypeDB = 1
	PayloadTypeContractBitcoinClient     MetricsPayloadTypeDB = 2
	PayloadTypeBitcoinNetwork            MetricsPayloadTypeDB = 3
	PayloadTypeContractTeleport          MetricsPayloadTypeDB = 4
	PayloadTypeContractCoordinator       MetricsPayloadTypeDB = 5
)

type MetricsPayloadDB struct {
	typeId  MetricsPayloadTypeDB
	payload string
}

type EventsPayloadDB struct {
	typeId  ?
	payload string
}
