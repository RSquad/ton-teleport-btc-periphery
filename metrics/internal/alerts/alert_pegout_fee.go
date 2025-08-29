package alerts

import (
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
)

type AlertPegoutFee struct {
	currentUnsignedPegout *coordinator.PegoutRecord
}

func NewAlertPegoutFee() Alert {
	return &AlertPegoutFee{
		currentUnsignedPegout: nil,
	}
}

func (alert *AlertPegoutFee) Check(dataSource AlertDataSource) (Severity, Labels, error) {
	labels := Labels{
		"bitcoin_tx_id": "",
		"pegout_addr":   "",
	}

	// Calulate severity
	severity := alert.GetSeverity()

	return severity, labels, nil
}

func (alert *AlertPegoutFee) GetSeverity() Severity {
	severity := SEVERITY_OK

	return severity
}
