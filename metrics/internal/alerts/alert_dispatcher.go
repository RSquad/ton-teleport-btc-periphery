package alerts

type AlertDispatcher interface {
	OnAlert(name string, labels []string, severity Severity)
}
