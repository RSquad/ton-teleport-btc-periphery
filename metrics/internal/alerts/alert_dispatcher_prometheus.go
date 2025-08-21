package alerts

type AlertDispatcherPrometheus struct {
}

func NewAlertDispatcherPrometheus() *AlertDispatcherPrometheus {
	dispatcher := AlertDispatcherPrometheus{}

	return &dispatcher
}

func (dispatcher *AlertDispatcherPrometheus) OnAlert(name string, severity Severity) {
	// TODO: implement
}
