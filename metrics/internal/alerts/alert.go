package alerts

import "fmt"

type (
	Severity    int
	Description string
	Values      map[string]any
)

const (
	SEVERITY_UNKNOWN  Severity = -1
	SEVERITY_OK       Severity = 0
	SEVERITY_INFO     Severity = 1
	SEVERITY_WARNING  Severity = 2
	SEVERITY_CRITICAL Severity = 3
)

type Alert interface {
	Check(dataSource AlertDataSource) (Severity, Description, Values, error)
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

var severityToString = map[Severity]string{
	SEVERITY_UNKNOWN:  "UNKNOWN",
	SEVERITY_OK:       "OK",
	SEVERITY_INFO:     "INFO",
	SEVERITY_WARNING:  "WARNING",
	SEVERITY_CRITICAL: "CRITICAL",
}

func (s Severity) String() string {
	if v, ok := severityToString[s]; ok {
		return v
	}
	return severityToString[SEVERITY_UNKNOWN]
}
