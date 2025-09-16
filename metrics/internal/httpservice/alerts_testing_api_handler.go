package httpservice

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/alerts"
)

type AlertsTestingApiHandler struct {
	alertManager *alerts.AlertManager
}

func NewAlertsTestingApiHandler(alertManager *alerts.AlertManager) *AlertsTestingApiHandler {
	return &AlertsTestingApiHandler{
		alertManager: alertManager,
	}
}

func (apiHandler AlertsTestingApiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check if 'action' parameter exists
	queryParams := r.URL.Query()
	actionName := queryParams.Get("action")
	if actionName == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Please set `action` argument"))
		return
	}

	var payload string
	var err error = nil

	switch actionName {
	case "info":
		payload, err = apiHandler.alertManager.GetInfoJson()
	case "enforce_info":
		payload, err = apiHandler.alertManager.GetEnforceInfoJsonStr()
	case "enforce_set":
		state, err := apiHandler.ParseState(&queryParams)
		if err != nil {
			break
		}

		err = apiHandler.alertManager.EnforceState(state)
		if err != nil {
			break
		}

		jsonData, err := json.Marshal(state)
		if err != nil {
			break
		}

		payload = string(jsonData)
	case "enforce_reset":
		state, err := apiHandler.ParseState(&queryParams)
		if err != nil {
			break
		}

		err = apiHandler.alertManager.ResetEnforceState(state.Name)
		if err != nil {
			break
		}

		jsonData, err := json.Marshal(state)
		if err != nil {
			break
		}

		payload = string(jsonData)
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Please select one of the next values: info, enforce_info, enforce_set, enforce_reset"))
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// Write data
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(payload))
}

func (apiHandler AlertsTestingApiHandler) ParseState(queryParams *url.Values) (*alerts.AlertState, error) {
	// name
	var nameStr string
	{
		nameStr = queryParams.Get("name")
		if nameStr == "" {
			return nil, errors.New("please set `name` argument")
		}
	}

	// severity
	var severity alerts.Severity = alerts.SEVERITY_UNKNOWN
	{
		severityStr := queryParams.Get("severity")

		var err error = nil
		if severityStr != "" {
			severity, err = alerts.StrToSeverity(severityStr)
			if err != nil {
				return nil, err
			}
		}
	}

	// labels
	var labels alerts.Labels = nil
	{
		labelsStr := queryParams.Get("labels")

		if labelsStr != "" {
			labels = make(alerts.Labels)
			values := strings.Split(labelsStr, "|")

			for _, v := range values {
				parts := strings.Split(v, "=")
				if len(parts) != 2 {
					return nil, fmt.Errorf("wrong label value '%s'", labelsStr)
				}

				labels[parts[0]] = parts[1]
			}
		}
	}

	// int_values
	var intValues alerts.IntValues = nil
	{
		intValuesStr := queryParams.Get("int_values")

		if intValuesStr != "" {
			intValues = make(alerts.IntValues)
			values := strings.Split(intValuesStr, "|")

			for _, v := range values {
				parts := strings.Split(v, "=")
				if len(parts) != 2 {
					return nil, fmt.Errorf("wrong value '%s'", intValuesStr)
				}

				n, err := strconv.ParseInt(parts[1], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("parse error: %w", err)
				}

				intValues[parts[0]] = n
			}
		}
	}

	return alerts.NewAlertState(nameStr, severity, labels, nil, true, intValues), nil
}
