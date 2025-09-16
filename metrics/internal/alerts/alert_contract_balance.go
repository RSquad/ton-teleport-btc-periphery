package alerts

import "github.com/xssnick/tonutils-go/address"

type AlertContractBalance struct {
	Name string
	Addr *address.Address
}

func NewAlertContractBalance(
	name string,
	addr *address.Address,
) Alert {
	return &AlertContractBalance{
		Name: name,
		Addr: addr,
	}
}

func (alert *AlertContractBalance) NewLabels() Labels {
	return Labels{
		"address": "",
	}
}

func (alert *AlertContractBalance) Check(dataSource AlertDataSource) (Severity, Labels, IntValues, error) {
	labels := alert.NewLabels()

	// Get current balance
	balance, err := dataSource.ActualContractBalance(alert.Name)
	if err != nil {
		return SEVERITY_UNKNOWN, labels, nil, err
	}

	severity := alert.GetSeverity(balance)
	labels["address"] = alert.Addr.StringRaw()

	return severity, labels, nil, nil
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
