package alerts

import (
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
	"github.com/xssnick/tonutils-go/address"
)

type AlertContractBalance struct {
	Name        string
	BalanceName string
	Addr        *address.Address
}

func NewAlertContractBalance(
	name string,
	balanceName string,
	addr *address.Address,
) Alert {
	return &AlertContractBalance{
		Name:        name,
		BalanceName: balanceName,
		Addr:        addr,
	}
}

func (alert *AlertContractBalance) Check(dataSource AlertDataSource) (Severity, Description, Values, error) {
	// Get current balance
	balance, err := dataSource.ActualContractBalance(alert.BalanceName)
	if err != nil {
		return SEVERITY_CRITICAL, "", nil, err
	}

	severity := alert.GetSeverity(balance)
	description := "OK"

	if severity > SEVERITY_OK {
		description = fmt.Sprintf(
			"The %s contract (%s) has a low balance: %s TON.",
			alert.BalanceName,
			mutils.TonExplorerLink(alert.Addr.StringRaw()),
			mutils.NanoIntToString(balance),
		)
	}

	return severity, Description(description), nil, nil
}

func (alert *AlertContractBalance) GetSeverity(balance int64) Severity {
	severity := SEVERITY_OK

	oneTon := int64(1000000000)

	if balance <= (oneTon / 2) {
		severity = SEVERITY_CRITICAL
	} else if balance <= (oneTon * 2) {
		severity = SEVERITY_WARNING
	}

	return severity
}
