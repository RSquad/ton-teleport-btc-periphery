package alerts

import (
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
)

type AlertTotalServiceFee struct{}

func NewAlertTotalServiceFee() Alert {
	return &AlertTotalServiceFee{}
}

func (alert *AlertTotalServiceFee) Check(dataSource AlertDataSource) (Severity, Description, Values, error) {
	// Get last signed pegout
	teleportContractStorage, err := dataSource.TeleportContractStorageDB()
	if err != nil {
		return SEVERITY_CRITICAL, "", nil, err
	}

	if teleportContractStorage == nil {
		return SEVERITY_OK, "OK", nil, nil
	}

	// Calulate severity
	severity := alert.GetSeverity(teleportContractStorage.TotalServiceFee)
	description := "OK"

	if severity > SEVERITY_OK {
		limit := 3000
		if severity == SEVERITY_CRITICAL {
			limit = 0
		}

		description = fmt.Sprintf(
			"Total service fee is less than %d satoshi. Steps to resolve: %s",
			limit,
			mutils.RunbookLink("TotalServiceFee"),
		)
	}

	return severity, Description(description), nil, nil
}

func (alert *AlertTotalServiceFee) GetSeverity(totalServiceFee int32) Severity {
	severity := SEVERITY_OK

	if totalServiceFee <= 0 {
		severity = SEVERITY_CRITICAL
	} else if totalServiceFee < 3000 {
		severity = SEVERITY_WARNING
	}

	return severity
}
