package events

import (
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
)

func (ed *EventDispatcher) formatParserNotFoundError(addr *address.Address) error {
	return fmt.Errorf("no parser found for address %s", utils.AddrToRawString(addr))
}
