package metrics

import "time"

type PayloadTypeDB int

const (
	PayloadTypeDKG PayloadTypeDB = iota
	PayloadTypePrevDKG
	PayloadTypeContractBitcoinClient
	PayloadTypeBlockChainInfo
	PayloadTypeContractTeleport
)

type PayloadDB struct {
	timestamp time.Time
	typeId    PayloadTypeDB
	payload   string
}
