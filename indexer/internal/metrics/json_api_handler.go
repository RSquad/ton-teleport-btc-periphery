package metrics

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
)

type JsonApiHandler struct {
	db                   *sql.DB
	tonClient            *tonclient.TonClient
	tonMaxMainValidators int
	cache                *Cache[string]
}

func NewJsonApiHandler(db *sql.DB, tonClient *tonclient.TonClient) *JsonApiHandler {
	return &JsonApiHandler{
		db:                   db,
		tonClient:            tonClient,
		tonMaxMainValidators: -1,
		cache:                NewCache[string](),
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
			payload, err = apiHandler.GetMints()
		case "burns":
			payload, err = apiHandler.GetBurns()
		case "reinits":
			payload, err = apiHandler.GetReinits()
		case "info":
			payload, err = apiHandler.GetInfo()
		case "internal_keys":
			payload, err = apiHandler.GetInternalKeys()
		case "plot_minted":
			payload, err = apiHandler.PlotMinted()
		case "plot_burned":
			payload, err = apiHandler.PlotBurned()
		case "plot_total_supply":
			payload, err = apiHandler.PlotTotalSupply()
		case "plots_summary":
			payload, err = apiHandler.GetPlotsSummary()
		case "dkg_status":
			payload, err = apiHandler.GetDkgStatus(r.Context())
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

func (apiHandler JsonApiHandler) GetMints() (string, error) {
	const limit = 5000 // Yes, we will select only last 5000 mints

	rows, err := apiHandler.db.Query(
		`SELECT COALESCE(json_agg(result), '[]') AS data FROM (
			SELECT
				m.created_at,
				m.status,
				TO_CHAR(m.amount::numeric(30,8) / 100000000::numeric(30,8), 'FM999999990.00000000') || ' BTC' AS amount,
				COALESCE(tt.hash, '_') AS ton_tx,
		    p.receiver_addr,
		    p.bitcoin_tx_id
	    FROM mints AS m
	    LEFT JOIN ton_txes AS tt ON m.ton_tx_mint = tt.id
	    LEFT JOIN pegins AS p ON m.id = p.mint_pegin
	    ORDER BY m.created_at DESC
	    LIMIT $1
		) AS result;`,
		limit,
	)

	if err != nil {
		return "", err
	}

	defer rows.Close()

	var data string
	if rows.Next() {
		err = rows.Scan(&data)
		if err != nil {
			return "", err
		}
	}

	if len(data) == 0 {
		data = "[]"
	}

	return data, nil
}

func (apiHandler JsonApiHandler) GetBurns() (string, error) {
	const limit = 5000 // Yes, we will select only last 5000 burns

	rows, err := apiHandler.db.Query(
		`SELECT COALESCE(json_agg(result), '[]') AS data FROM (
			SELECT
				tt.created_at,
				TO_CHAR(b.amount::numeric(30,8) / 100000000::numeric(30,8), 'FM999999990.00000000') || ' BTC' AS amount,
				COALESCE(p.addr, '_') AS pegout_addr,
				b.sender_addr,
				COALESCE(p.bitcoin_tx_id, '_') AS bitcoin_tx_id,
				COALESCE(p.bitcoin_tx_raw, '_') AS bitcoin_tx_raw,
				COALESCE(tt.hash, '_') AS ton_tx,
				COALESCE(p.status, '_') AS pegout_status 
			FROM burns AS b
			LEFT JOIN ton_txes AS tt ON tt.id = b.ton_tx_burn
			LEFT JOIN pegouts AS p ON p.id = b.pegout_burn
			WHERE b.sender_addr != ':0'
			ORDER BY tt.created_at DESC 
			LIMIT $1
		) AS result;`,
		limit,
	)
	if err != nil {
		return "", err
	}

	defer rows.Close()

	var data string
	if rows.Next() {
		err = rows.Scan(&data)
		if err != nil {
			return "", err
		}
	}

	if len(data) == 0 {
		data = "[]"
	}

	return data, nil
}

func (apiHandler JsonApiHandler) GetReinits() (string, error) {
	const limit = 5000 // Yes, we will select only last 5000 reinits

	rows, err := apiHandler.db.Query(
		`SELECT COALESCE(json_agg(result), '[]') AS data FROM (
		  SELECT
		    tt.created_at AS created_at,
				tt.hash AS ton_tx,
		    TO_CHAR(r.amount::numeric(30,8) / 100000000::numeric(30,8), 'FM999999990.00000000') || ' BTC' AS amount,
		    COALESCE(p.addr, '_') AS pegout_addr,
		    COALESCE(p.bitcoin_tx_id, '_') AS bitcoin_tx_id,
				COALESCE(p.bitcoin_tx_raw, '_') AS bitcoin_tx_raw,
		    COALESCE(p.status, '_') AS pegout_status
		  FROM ton_txes AS tt
			INNER JOIN reinits AS r ON tt.id = r.ton_tx_reinit
		  LEFT JOIN pegouts AS p ON p.id = r.pegout_reinit
		  ORDER BY created_at DESC
		  LIMIT $1
		) AS result;`,
		limit,
	)
	if err != nil {
		return "", err
	}

	defer rows.Close()

	var data string
	if rows.Next() {
		err = rows.Scan(&data)
		if err != nil {
			return "", err
		}
	}

	if len(data) == 0 {
		data = "[]"
	}

	return data, nil
}

func (apiHandler JsonApiHandler) GetInternalKeys() (string, error) {
	const limit = 5000 // Yes, we will select only last 5000 internal keys

	rows, err := apiHandler.db.Query(
		`SELECT COALESCE(json_agg(result), '[]') AS data FROM (
		  SELECT 
			  ik.completed_at,
				ik.key AS internal_key,
				tt.hash AS ton_tx
			FROM internal_keys AS ik
			JOIN ton_txes AS tt ON tt.id = ik.ton_tx_internal_key
			ORDER BY ik.id DESC
			LIMIT $1
		) AS result;`,
		limit,
	)
	if err != nil {
		return "", err
	}

	defer rows.Close()

	var data string
	if rows.Next() {
		err = rows.Scan(&data)
		if err != nil {
			return "", err
		}
	}

	if len(data) == 0 {
		data = "[]"
	}

	return data, nil
}

func (apiHandler JsonApiHandler) GetInfo() (string, error) {
	rows, err := apiHandler.db.Query(
		`SELECT jsonb_build_object(
				'contractBitcoinClient', (
						SELECT payload::json
						FROM metrics_data
						WHERE type_id = 2
						ORDER BY id DESC
						LIMIT 1
				),
				'blockChainInfo', (
						SELECT payload::json
						FROM metrics_data
						WHERE type_id = 3
						ORDER BY id DESC
						LIMIT 1
				),
				'contractTeleport', (
						SELECT payload::json
						FROM metrics_data
						WHERE type_id = 4
						ORDER BY id DESC
						LIMIT 1
				)
		) AS result;`,
	)
	if err != nil {
		return "", err
	}

	defer rows.Close()

	var data string
	if rows.Next() {
		err = rows.Scan(&data)
		if err != nil {
			return "", err
		}
	}

	if len(data) == 0 {
		data = "{}"
	}

	return data, nil
}

func (apiHandler JsonApiHandler) PlotMinted() (string, error) {
	rows, err := apiHandler.db.Query(
		`SELECT COALESCE(json_agg(result), '[]') AS data FROM (
			WITH data_by_days AS (
				SELECT
					DATE_TRUNC('day', created_at)::date AS day,
					SUM(amount::int8) AS minted,
					COUNT(1) AS count
				FROM mints
				WHERE status = 'SUCCESS'
				GROUP BY DATE_TRUNC('day', created_at)
			) SELECT day, TO_CHAR(minted::numeric(30,8) / 100000000::numeric(30,8), 'FM999999990.00000000') AS minted, count FROM data_by_days ORDER BY day ASC
		) AS result;`,
	)
	if err != nil {
		return "", err
	}

	defer rows.Close()

	var data string
	if rows.Next() {
		err = rows.Scan(&data)
		if err != nil {
			return "", err
		}
	}

	if len(data) == 0 {
		data = "[]"
	}

	return data, nil
}

func (apiHandler JsonApiHandler) PlotBurned() (string, error) {
	rows, err := apiHandler.db.Query(
		`SELECT COALESCE(json_agg(result), '[]') AS data FROM (
			WITH data_by_days AS (
				SELECT
					DATE_TRUNC('day', tt.created_at)::date AS day,
					SUM(b.amount::int8) AS burned,
					COUNT(1) AS count
				FROM burns AS b 
				JOIN ton_txes AS tt ON tt.id = b.ton_tx_burn
				JOIN pegouts AS p ON p.id = b.pegout_burn 
				WHERE p.status = 'CONFIRMED' AND b.sender_addr != ':0'
				GROUP BY DATE_TRUNC('day', tt.created_at)
  		) SELECT day, TO_CHAR(burned::numeric(30,8) / 100000000::numeric(30,8), 'FM999999990.00000000') AS burned, count FROM data_by_days ORDER BY day ASC
		) AS result;`,
	)
	if err != nil {
		return "", err
	}

	defer rows.Close()

	var data string
	if rows.Next() {
		err = rows.Scan(&data)
		if err != nil {
			return "", err
		}
	}

	if len(data) == 0 {
		data = "[]"
	}

	return data, nil
}

func (apiHandler JsonApiHandler) PlotTotalSupply() (string, error) {
	rows, err := apiHandler.db.Query(
		`SELECT COALESCE(json_agg(result), '[]') AS data FROM (
			WITH unified_events AS (
				SELECT
					DATE_TRUNC('day', created_at)::date AS day,
					amount::int8 AS value
				FROM mints
				WHERE status = 'SUCCESS'

				UNION ALL

				SELECT
					DATE_TRUNC('day', tt.created_at)::date AS day, 
					-b.amount::int8 AS value
				FROM burns AS b
				JOIN ton_txes AS tt ON tt.id = b.ton_tx_burn
				JOIN pegouts AS p ON p.id = b.pegout_burn
				WHERE p.status = 'CONFIRMED' AND b.sender_addr != ':0'
			),

			daily_totals AS (
				SELECT
					day,
					SUM(value) AS daily_sum
				FROM unified_events
				GROUP BY day
			)

			SELECT
				day,
				SUM(daily_sum::numeric(30,8) / 100000000::numeric(30,8)) OVER (ORDER BY day) AS cumulative_total
			FROM daily_totals
			ORDER BY day
		) AS result;`,
	)
	if err != nil {
		return "", err
	}

	defer rows.Close()

	var data string
	if rows.Next() {
		err = rows.Scan(&data)
		if err != nil {
			return "", err
		}
	}

	if len(data) == 0 {
		data = "[]"
	}

	return data, nil
}

func (apiHandler JsonApiHandler) GetPlotsSummary() (string, error) {
	rows, err := apiHandler.db.Query(
		`SELECT jsonb_build_object(
				'mints_count', (
						SELECT COUNT(1) AS row_count FROM mints WHERE status = 'SUCCESS'
				),
				'burns_count', (
						SELECT COUNT(1) AS row_count FROM burns AS b INNER JOIN pegouts AS p ON b.pegout_burn = p.id AND p.status = 'CONFIRMED' AND b.sender_addr != ':0'
				),
				'total_minted', (
						SELECT COALESCE(SUM(amount::int8)::numeric(30,8) / 100000000::numeric(30,8), 0) AS total_minted FROM mints WHERE status = 'SUCCESS'
				),
				'total_burned', (
						SELECT COALESCE(SUM(b.amount::int8)::numeric(30,8) / 100000000::numeric(30,8), 0) AS total_burned FROM burns AS b JOIN pegouts AS p ON p.id = b.pegout_burn WHERE p.status = 'CONFIRMED' AND b.sender_addr != ':0'
				)
		) AS result;`,
	)
	if err != nil {
		return "", err
	}

	defer rows.Close()

	var data string
	if rows.Next() {
		err = rows.Scan(&data)
		if err != nil {
			return "", err
		}
	}

	if len(data) == 0 {
		data = "{}"
	}

	return data, nil
}

func (apiHandler JsonApiHandler) GetDkgStatus(ctx context.Context) (string, error) {
	type OriginalData struct {
		Dkg     map[string]interface{}
		PrevDkg map[string]interface{}
	}

	type DkgInfo struct {
		VSetSize                int
		ValidatorsCountInDkg    int
		ValidatorsCountNotInDkg int
		ValidatorsCountMax      int
		ValidatorsCountTotal    int
		ValidatorsIdxInDkg      map[int]string
		ValidatorsIdxNotInDkg   map[int]string
	}

	type DkgStatus struct {
		Dkg      DkgInfo
		PrevDkg  DkgInfo
		Original OriginalData
	}

	// Select DKG
	dkg, err := apiHandler.SelectToObject("SELECT payload FROM metrics_data WHERE type_id = 0 ORDER BY id DESC LIMIT 1")
	if err != nil {
		return "", err
	}

	// Select prevDKG
	prevDkg, err := apiHandler.SelectToObject("SELECT payload FROM metrics_data WHERE type_id = 1 ORDER BY id DESC LIMIT 1")
	if err != nil {
		return "", err
	}

	var status DkgStatus

	// Fill DkgStatus
	{
		status.Original.Dkg = dkg
		status.Original.PrevDkg = prevDkg

		fillFn := func(dkg map[string]interface{}) (*DkgInfo, error) {
			var dkgInfo DkgInfo

			vset, ok := dkg["VSet"].(map[string]interface{})
			if !ok {
				return nil, errors.New("VSet has wrong type")
			}

			maxSignersJson, ok := dkg["MaxSigners"].(json.Number)
			if !ok {
				return nil, errors.New("MaxSigners has wrong type")
			}

			var dkgLastRound *map[string]interface{} = nil
			roundNames := []string{"R3", "R2", "R1"}
			for _, roundName := range roundNames {
				if _, exists := dkg[roundName]; exists {
					res, ok := dkg[roundName].(map[string]interface{})
					if !ok {
						return nil, fmt.Errorf("`%s` has wrong type", roundName)
					}

					dkgLastRound = &res
					break
				}
			}

			maxSigners, err := maxSignersJson.Int64()
			if err != nil {
				return nil, err
			}

			// tonMaxMainValidators
			if apiHandler.tonMaxMainValidators < 0 {
				block, err := apiHandler.tonClient.API.GetMasterchainInfo(ctx)
				if err != nil {
					return nil, fmt.Errorf("failed to get block: %v", err)
				}

				tonConfig, err := apiHandler.tonClient.API.GetBlockchainConfig(ctx, block, 16)
				if err != nil {
					return nil, fmt.Errorf("failed to get config: %v", err)
				}

				tonConfigParam16 := tonConfig.Get(16)
				s := tonConfigParam16.BeginParse()
				s.MustLoadUInt(16)
				apiHandler.tonMaxMainValidators = int(s.MustLoadUInt(16))
			}

			dkgInfo.ValidatorsCountMax = apiHandler.tonMaxMainValidators
			dkgInfo.VSetSize = len(vset)
			dkgInfo.ValidatorsCountTotal = dkgInfo.ValidatorsCountMax
			dkgInfo.ValidatorsCountInDkg = int(maxSigners)
			if len(vset) < dkgInfo.ValidatorsCountTotal {
				dkgInfo.ValidatorsCountTotal = len(vset)
			}
			dkgInfo.ValidatorsCountNotInDkg = dkgInfo.ValidatorsCountTotal - dkgInfo.ValidatorsCountInDkg

			if dkgLastRound != nil {
				maskJson, ok := (*dkgLastRound)["Mask"].(json.Number)
				if !ok {
					return nil, errors.New("mask has wrong type")
				}

				mask := new(big.Int)
				mask, ok = mask.SetString(maskJson.String(), 10)
				if !ok {
					return nil, errors.New("invalid bigint")
				}

				dkgInfo.ValidatorsIdxInDkg = make(map[int]string)
				dkgInfo.ValidatorsIdxNotInDkg = make(map[int]string)

				for i := 0; i < dkgInfo.ValidatorsCountTotal; i++ {
					pubKeyBase64, ok := vset[strconv.Itoa(i)].(string)
					if !ok {
						return nil, errors.New("invalid vset pubkey type")
					}

					if mask.Bit(i) == 1 {
						dkgInfo.ValidatorsIdxInDkg[i] = pubKeyBase64
					} else {
						dkgInfo.ValidatorsIdxNotInDkg[i] = pubKeyBase64
					}
				}
			}

			return &dkgInfo, nil
		}

		statusDkg, err := fillFn(status.Original.Dkg)
		if err != nil {
			return "", err
		}
		status.Dkg = *statusDkg

		statusPrevDkg, err := fillFn(status.Original.PrevDkg)
		if err != nil {
			return "", err
		}
		status.PrevDkg = *statusPrevDkg
	}

	jsonData, err := json.Marshal(status)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

func (apiHandler JsonApiHandler) SelectToObject(sql string) (map[string]interface{}, error) {
	rows, err := apiHandler.db.Query(sql)
	if err != nil {
		return nil, err
	}

	var data string
	if rows.Next() {
		err = rows.Scan(&data)
		if err != nil {
			return nil, err
		}
	}

	if len(data) == 0 {
		data = "{}"
	}

	jsonDec := json.NewDecoder(strings.NewReader(data))
	jsonDec.UseNumber()

	var m map[string]interface{}
	if err := jsonDec.Decode(&m); err != nil {
		return nil, err
	}
	return m, nil
}
