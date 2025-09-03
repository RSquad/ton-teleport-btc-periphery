package alerts

import (
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
)

type AlertContractBalance struct {
	Name string
}

func NewAlertContractBalance(name string) Alert {
	return &AlertContractBalance{
		Name: name,
	}
}

func (alert *AlertContractBalance) Check(dataSource AlertDataSource) (Severity, Labels, error) {
	labels := Labels{
		"address": "",
	}

	// Get current balance
	balances, err := dataSource.ActualContractBalances()
	if err != nil {
		return SEVERITY_UNKNOWN, labels, err
	}

	// Find balance by name
	balance, err := alert.FindBalance(alert.Name, balances)
	if err != nil {
		return SEVERITY_UNKNOWN, labels, err
	}

	severity := alert.GetSeverity(balance)
	labels["address"] = balance.Addr.StringRaw()

	return severity, labels, nil
}

func (alert *AlertContractBalance) FindBalance(
	name string,
	balances *data_models.ContractBalances,
) (*data_models.ContractBalance, error) {
	for _, balance := range balances.Balances {
		if balance.Name == name {
			return balance, nil
		}
	}
	return nil, fmt.Errorf("balance not found for name '%s'", name)
}

func (alert *AlertContractBalance) GetSeverity(balance *data_models.ContractBalance) Severity {
	severity := SEVERITY_OK

	oneTon := uint64(1000000000)

	if balance.Balance <= (oneTon / 2) {
		severity = SEVERITY_CRITICAL
	} else if balance.Balance <= (oneTon * 2) {
		severity = SEVERITY_WARNING
	}

	return severity
}
