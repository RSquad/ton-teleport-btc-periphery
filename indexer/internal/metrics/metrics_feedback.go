package metrics

import (
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
)

func (m *Metrics) formatGetBalanceError(addr *address.Address) error {
	return fmt.Errorf("can not get balance for contract: %v", utils.AddrToRawString(addr))
}
