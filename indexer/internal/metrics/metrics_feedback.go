package metrics

import (
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
)

func (m *Metrics) formatGetBalanceError(addr *address.Address) error {
	return fmt.Errorf("can not get balance for contract: %v", utils.AddrToRawString(addr))
}

func (m *Metrics) formatParseFloatError(s string) error {
	return fmt.Errorf("can not convert string %v to float", s)
}

func (m *Metrics) formatGetLastConfirmedBlockHashError() error {
	return fmt.Errorf("can not get last confirmed block hash")
}

func (m *Metrics) formatGetCandidateBlockHashesError() error {
	return fmt.Errorf("can not get candidate block hashes")
}

func (m *Metrics) formatGetConfirmationsNeededError() error {
	return fmt.Errorf("can not get confirmations needed")
}

func (m *Metrics) formatGetLastConfirmedBlockHeightError() error {
	return fmt.Errorf("can not get last confirmed block hash")
}