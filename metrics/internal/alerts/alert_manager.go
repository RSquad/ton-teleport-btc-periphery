package alerts

import (
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
)

type AlertManager struct {
	alerts          map[string]Alert
	dataSource      AlertDataSource
	alertDispatcher AlertDispatcher
}

func NewAlertManager(
	dataSource AlertDataSource,
	alertDispatcher AlertDispatcher,
) *AlertManager {
	alertManager := AlertManager{
		alerts:          make(map[string]Alert),
		dataSource:      dataSource,
		alertDispatcher: alertDispatcher,
	}

	return &alertManager
}

func (manager *AlertManager) RegisterAlert(name string, alert Alert) error {
	if _, exists := manager.alerts[name]; exists {
		return fmt.Errorf("Alert with name `%s` already exists", name)
	}

	manager.alerts[name] = alert

	return nil
}

func (manager *AlertManager) CheckAll() {
	for alertName, alert := range manager.alerts {
		severity, err := alert.Check(manager.dataSource)
		if err != nil {
			manager.LogAlertError(alertName, err)
			continue
		}

		if severity >= 0 {
			manager.alertDispatcher.OnAlert(alertName, severity)
		}
	}
}

func (manager *AlertManager) LogAlertError(alertName string, err error) {
	logger.Log.Error().
		Str("Alert", alertName).
		Err(err).
		Msg("Alert finished work with error")
}
