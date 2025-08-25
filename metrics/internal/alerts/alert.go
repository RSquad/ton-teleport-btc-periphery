package alerts

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
