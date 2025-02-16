package events

import (
	"fmt"
	"strings"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
)

func (ed *EventDispatcher) handleTonTxWriteError(err error) (bool, error) {
	if err == nil {
		return true, nil
	}
	if strings.Contains(err.Error(), "duplicate key value violates unique constraint \"ton_txes_hash_key\"") {
		return false, nil
	}
	return false, err
}

func (ed *EventDispatcher) formatParserNotFoundError(addr *address.Address) error {
	return fmt.Errorf("no parser found for address %s", utils.AddrToRawString(addr))
}
