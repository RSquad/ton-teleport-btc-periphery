package data_models

import (
	"fmt"

	"github.com/xssnick/tonutils-go/address"
)

type PegoutStatus int

const (
	PEGOUT_CONFIRMED PegoutStatus = iota
	PEGOUT_SIGNED
	PEGOUT_SIGNING
)

type PegoutDbRow struct {
	Id               uint64
	Addr             *address.Address
	Status           PegoutStatus
	BitcoinTxRaw     []byte
	BitcoinTxId      []byte
	BitcoinBlockHash []byte
}

func PegoutStatusFromString(str string) (PegoutStatus, error) {
	switch str {
	case "CONFIRMED":
		return PEGOUT_CONFIRMED, nil
	case "SIGNED":
		return PEGOUT_SIGNED, nil
	case "SIGNING":
		return PEGOUT_SIGNING, nil
	}

	return 0, fmt.Errorf("failed to convert `%s` to PegoutStatus", str)
}
