package httpservice

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/metrics"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/utils"
)

type JsonApiHandler struct {
	metricsManager *metrics.MetricsManager
	cache          *utils.Cache[string]
}

func NewJsonApiHandler(db *sql.DB, tonClient *tonclient.TonClient) *JsonApiHandler {
	return &JsonApiHandler{
		metricsManager: metrics.NewMetricsManager(db, tonClient),
		cache:          utils.NewCache[string](),
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

	cachedValue, ok := apiHandler.cache.Get(sourceName)

	if ok {
		payload = cachedValue
	} else {
		switch sourceName {
		case "mints":
			payload, err = apiHandler.metricsManager.GetMints()
		case "burns":
			payload, err = apiHandler.metricsManager.GetBurns()
		case "reinits":
			payload, err = apiHandler.metricsManager.GetReinits()
		case "info":
			payload, err = apiHandler.metricsManager.GetInfo()
		case "internal_keys":
			payload, err = apiHandler.metricsManager.GetInternalKeys()
		case "plot_minted":
			payload, err = apiHandler.metricsManager.PlotMinted()
		case "plot_burned":
			payload, err = apiHandler.metricsManager.PlotBurned()
		case "plot_total_supply":
			payload, err = apiHandler.metricsManager.PlotTotalSupply()
		case "plots_summary":
			payload, err = apiHandler.metricsManager.GetPlotsSummary()
		case "dkg_status":
			payload, err = apiHandler.metricsManager.GetDkgStatus(r.Context())
		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Please select one of the next values: mints, burns, reinits, info, internal_keys, plot_minted, plot_burned, plot_total_supply, plots_summary, dkg_status"))
			return
		}

		apiHandler.cache.Set(sourceName, payload, 30*time.Second)
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
