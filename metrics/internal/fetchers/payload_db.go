package fetchers

import "time"

type PayloadTypeDB int

const (
	PayloadTypeDKG                   PayloadTypeDB = iota
	PayloadTypePrevDKG               PayloadTypeDB = 1
	PayloadTypeContractBitcoinClient PayloadTypeDB = 2
	PayloadTypeBitcoinNetwork        PayloadTypeDB = 3
	PayloadTypeContractTeleport      PayloadTypeDB = 4
	PayloadTypeContractCoordinator   PayloadTypeDB = 5
)

type PayloadDB struct {
	timestamp time.Time
	typeId    PayloadTypeDB
	payload   string
}
