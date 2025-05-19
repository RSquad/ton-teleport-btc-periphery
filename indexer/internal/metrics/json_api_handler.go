package metrics

import (
	"database/sql"
	"net/http"
	"time"
)

type JsonApiHandler struct {
	db    *sql.DB
	cache *Cache
}

func NewJsonApiHandler(db *sql.DB) *JsonApiHandler {
	return &JsonApiHandler{
		db:    db,
		cache: NewCache(),
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
		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Please select one of the next values: mints, burns, reinits, info"))
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
		`SELECT json_agg(result) AS data FROM (
			SELECT
				m.created_at,
				m.status,
				TO_CHAR(CAST(m.amount AS real) / 100000000.0, 'FM999999990.00000000') || ' BTC' AS amount,
				COALESCE(tt.hash, '_') AS ton_tx,
		    p.receiver_addr,
		    p.bitcoin_tx_id
	    FROM mints AS m
	    LEFT JOIN ton_txes AS tt ON m.ton_tx_mint = tt.id
	    LEFT JOIN pegins AS p ON m.id = p.mint_pegin
	    ORDER BY m.id DESC
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

	return data, nil
}

func (apiHandler JsonApiHandler) GetBurns() (string, error) {
	const limit = 5000 // Yes, we will select only last 5000 burns

	rows, err := apiHandler.db.Query(
		`SELECT json_agg(result) AS data FROM (
			SELECT
				tt.created_at,
				TO_CHAR(CAST(b.amount AS real) / 100000000.0, 'FM999999990.00000000') || ' BTC' AS amount,
				COALESCE(p.addr, '_') AS pegout_addr,
				b.sender_addr,
				COALESCE(p.bitcoin_tx_id, '_') AS bitcoin_tx_id,
				COALESCE(tt.hash, '_') AS ton_tx,
				COALESCE(p.status, '_') AS pegout_status 
			FROM burns AS b 
			LEFT JOIN ton_txes AS tt ON tt.id = b.ton_tx_burn
			LEFT JOIN pegouts AS p ON p.id = b.pegout_burn 
			ORDER BY b.id DESC 
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

	return data, nil
}

func (apiHandler JsonApiHandler) GetReinits() (string, error) {
	const limit = 5000 // Yes, we will select only last 5000 reinits

	rows, err := apiHandler.db.Query(
		`SELECT json_agg(result) AS data FROM (
		  SELECT
		    tt.created_at AS created_at,
				tt.hash AS ton_tx,
		    TO_CHAR(CAST(r.amount AS real) / 100000000.0, 'FM999999990.00000000') || ' BTC' AS amount,
		    COALESCE(p.addr, '_') AS pegout_addr,
		    COALESCE(p.bitcoin_tx_id, '_') AS bitcoin_tx_id,		    
		    COALESCE(p.status, '_') AS pegout_status
		  FROM ton_txes AS tt
			INNER JOIN reinits AS r ON tt.id = r.ton_tx_reinit
		  LEFT JOIN pegouts AS p ON p.id = r.pegout_reinit
		  ORDER BY r.id DESC
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

	return data, nil
}

func (apiHandler JsonApiHandler) GetInfo() (string, error) {
	// SELECT payload FROM metrics_data WHERE type_id = 2 ORDER BY id DESC LIMIT 1;
	// {"CandidateBlockHashes":["0000000ddcbcd1b600cb8d21ca8a4c53441b966e691fa1bfa8c3effd09de464c","00000005fe81d7938574732c03da649a76f00a2d19e1aeacf67b38ad45430666"],"LastConfirmedBlockHash":"000000117f800c3982612967561b8a08174082c27a81dc2e89d2387b6afe8311","ConfirmationsNeeded":2,"LastConfirmedBlockHeight":252284}

	// SELECT payload FROM metrics_data WHERE type_id = 3 ORDER BY id DESC LIMIT 1;
	// {"chain":"signet","blocks":252286,"bestblockhash":"0000000ddcbcd1b600cb8d21ca8a4c53441b966e691fa1bfa8c3effd09de464c","mediantime":1747383525}

	rows, err := apiHandler.db.Query(
		`SELECT AS data FROM (
		  SELECT
		    tt.created_at AS created_at,
				tt.hash AS ton_tx,
		    TO_CHAR(CAST(r.amount AS real) / 100000000.0, 'FM999999990.00000000') || ' BTC' AS amount,
		    COALESCE(p.addr, '_') AS pegout_addr,
		    COALESCE(p.bitcoin_tx_id, '_') AS bitcoin_tx_id,		    
		    COALESCE(p.status, '_') AS pegout_status
		  FROM ton_txes AS tt
			INNER JOIN reinits AS r ON tt.id = r.ton_tx_reinit
		  LEFT JOIN pegouts AS p ON p.id = r.pegout_reinit
		  ORDER BY r.id DESC
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

	return data, nil
}
