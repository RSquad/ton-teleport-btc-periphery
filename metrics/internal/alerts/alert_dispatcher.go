package alerts

type AlertDispatcher interface {
	OnAlert(state *AlertState)
}
