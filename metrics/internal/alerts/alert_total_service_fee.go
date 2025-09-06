package alerts

type AlertTotalServiceFee struct{}

func NewAlertTotalServiceFee() Alert {
	return &AlertTotalServiceFee{}
}

func (alert *AlertTotalServiceFee) Check(dataSource AlertDataSource) (Severity, Labels, IntValues, error) {
	// Get last signed pegout
	teleportContractStorage, err := dataSource.TeleportContractStorageDB()
	if err != nil {
		return SEVERITY_UNKNOWN, nil, nil, err
	}

	if teleportContractStorage == nil {
		return SEVERITY_OK, nil, nil, nil
	}

	// Calulate severity
	severity := alert.GetSeverity(teleportContractStorage.TotalServiceFee)

	return severity, nil, nil, nil
}

func (alert *AlertTotalServiceFee) GetSeverity(totalServiceFee int32) Severity {
	severity := SEVERITY_OK

	if totalServiceFee <= 0 {
		severity = SEVERITY_CRITICAL
	} else if totalServiceFee <= 1000 {
		severity = SEVERITY_WARNING
	}

	return severity
}
