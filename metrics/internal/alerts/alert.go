package alerts

import "fmt"

type Severity int
type Labels map[string]string
type IntValues map[string]int64

const (
	SEVERITY_UNKNOWN  Severity = -1
	SEVERITY_OK       Severity = 0
	SEVERITY_INFO     Severity = 1
	SEVERITY_WARNING  Severity = 2
	SEVERITY_CRITICAL Severity = 3
)

type Alert interface {
	Check(dataSource AlertDataSource) (Severity, Labels, IntValues, error)
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
