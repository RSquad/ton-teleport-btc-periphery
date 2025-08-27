package alerts

import "fmt"

type Severity int

const (
	SEVERITY_UNKNOWN  Severity = -2
	SEVERITY_OK       Severity = -1
	SEVERITY_INFO     Severity = 0
	SEVERITY_WARNING  Severity = 1
	SEVERITY_CRITICAL Severity = 2
)

type Alert interface {
	Check(dataSource AlertDataSource) (Severity, []string, error)
}

func StrToSeverity(s string) (Severity, error) {
	severity := SEVERITY_UNKNOWN

	switch s {
	case "UNKNOWN":
		severity = SEVERITY_UNKNOWN
	case "OK":
		severity = SEVERITY_OK
	case "INFO":
		severity = SEVERITY_INFO
	case "WARNING":
		severity = SEVERITY_WARNING
	case "CRITICAL":
		severity = SEVERITY_CRITICAL
	default:
		return SEVERITY_UNKNOWN, fmt.Errorf("unknown severity value `%s`", s)
	}

	return severity, nil
}
