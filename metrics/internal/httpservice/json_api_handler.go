package httpservice

import (
	"net/http"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/alerts"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/metrics"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
)

type JsonApiHandler struct {
	metricsManager *metrics.MetricsManager
	alertsManager  *alerts.AlertManager
	cache          *mutils.Cache[string]
}

func NewJsonApiHandler(
	metricsManager *metrics.MetricsManager,
	alertsManager *alerts.AlertManager,
) *JsonApiHandler {
	return &JsonApiHandler{
		metricsManager: metricsManager,
		alertsManager:  alertsManager,
		cache:          mutils.NewCache[string](),
	}
}

func (apiHandler JsonApiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check if 'source' parameter exists
	queryParams := r.URL.Query()
	sourceName := queryParams.Get("source")
	if sourceName == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Please set `source` argument"))
		return
	}

	var payload string
	var err error = nil

	cachedValue, ok := apiHandler.cache.Get(queryParams.Encode())

	if ok {
		payload = cachedValue
	} else {
		switch sourceName {
		case "mints":
			payload, err = apiHandler.metricsManager.MintsJson()
		case "burns":
			payload, err = apiHandler.metricsManager.BurnsJson()
		case "reinits":
			payload, err = apiHandler.metricsManager.ReinitsJson()
		case "info":
			payload, err = apiHandler.metricsManager.InfoJson()
		case "internal_keys":
			payload, err = apiHandler.metricsManager.InternalKeysJson()
		case "plot_minted":
			payload, err = apiHandler.metricsManager.PlotMintedJson()
		case "plot_burned":
			payload, err = apiHandler.metricsManager.PlotBurnedJson()
		case "plot_total_supply":
			payload, err = apiHandler.metricsManager.PlotTotalSupplyJson()
		case "plots_summary":
			payload, err = apiHandler.metricsManager.PlotsSummaryJson()
		case "dkg_status":
			payload, err = apiHandler.metricsManager.DkgStatusJson(r.Context())
		case "alerts":
			payload, err = apiHandler.alertsManager.GetInfoJson()
		case "system_info":
			payload, err = apiHandler.metricsManager.SystemInfoJson()
		case "contract_balances":
			name := queryParams.Get("name")
			if name == "" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Please set `name` argument"))
				return
			}

			payload, err = apiHandler.metricsManager.ContractBalanceJson(name)
		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Please select one of the next values: mints, burns, reinits, info, internal_keys, plot_minted, plot_burned, plot_total_supply, plots_summary, dkg_status, alerts, system_info, contract_balances"))
			return
		}

		if len(payload) == 0 {
			payload = "{}"
		}

		apiHandler.cache.Set(queryParams.Encode(), payload, 30*time.Second)
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
