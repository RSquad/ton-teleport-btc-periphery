package alerts

type AlertDispatcher interface {
	OnAlert(name string, severity Severity)
}
