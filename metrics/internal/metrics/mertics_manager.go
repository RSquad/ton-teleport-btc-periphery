package metrics

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"math/big"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/alerts"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/config"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_sources"
	"github.com/xssnick/tonutils-go/address"
)

type MetricsManager struct {
	db                  *sql.DB
	dataSourceDB        *data_sources.DataSourceDB
	globalRuntimeConfig *config.GlobalRuntimeConfig
	contractAddrs       map[string]*address.Address
	alertManager        *alerts.AlertManager
}

func NewMetricsManager(
	db *sql.DB,
	globalRuntimeConfig *config.GlobalRuntimeConfig,
	contractAddrs map[string]*address.Address,
	alertManager *alerts.AlertManager,
) *MetricsManager {
	return &MetricsManager{
		db:                  db,
		dataSourceDB:        data_sources.NewDataSourceDB(db),
		globalRuntimeConfig: globalRuntimeConfig,
		contractAddrs:       contractAddrs,
		alertManager:        alertManager,
	}
}

func (manager *MetricsManager) MintsJson() (string, error) {
	const limit = 5000 // Yes, we will select only last 5000 mints

	rows, err := manager.db.Query(
		`SELECT COALESCE(jsonb_agg(result), '[]') AS data FROM (
			SELECT
				m.created_at,
				m.status,
				TO_CHAR(m.amount::numeric(24,8) / 100000000::numeric(24,8), 'FM999999990.00000000') || ' BTC' AS amount,
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

func (manager *MetricsManager) BurnsJson() (string, error) {
	const limit = 5000 // Yes, we will select only last 5000 burns

	rows, err := manager.db.Query(
		`SELECT COALESCE(jsonb_agg(result), '[]') AS data FROM (
			SELECT
				tt.created_at,
				TO_CHAR(b.amount::numeric(24,8) / 100000000::numeric(24,8), 'FM999999990.00000000') || ' BTC' AS amount,
				COALESCE(p.addr, '_') AS pegout_addr,
				b.sender_addr,
				COALESCE(p.bitcoin_tx_id, '_') AS bitcoin_tx_id,
				COALESCE(p.bitcoin_tx_raw, '_') AS bitcoin_tx_raw,
				COALESCE(tt.hash, '_') AS ton_tx,
				COALESCE(p.status, '_') AS pegout_status 
			FROM burns AS b
			LEFT JOIN ton_txes AS tt ON tt.id = b.ton_tx_burn
			LEFT JOIN pegouts AS p ON p.id = b.pegout_burn
			WHERE b.sender_addr != '0:'
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

func (manager *MetricsManager) ReinitsJson() (string, error) {
	const limit = 5000 // Yes, we will select only last 5000 reinits

	rows, err := manager.db.Query(
		`SELECT COALESCE(jsonb_agg(result), '[]') AS data FROM (
		  SELECT
		    tt.created_at AS created_at,
				tt.hash AS ton_tx,
		    TO_CHAR(r.amount::numeric(24,8) / 100000000::numeric(24,8), 'FM999999990.00000000') || ' BTC' AS amount,
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

func (manager *MetricsManager) InternalKeysJson() (string, error) {
	const limit = 5000 // Yes, we will select only last 5000 internal keys

	rows, err := manager.db.Query(
		`SELECT COALESCE(jsonb_agg(result), '[]') AS data FROM (
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

func (manager *MetricsManager) InfoJson() (string, error) {
	type ContractTeleportUTXO struct {
		Address       string
		Amount        *big.Int
		Index         uint8
		TapMerkleRoot *chainhash.Hash
		MintAddress   string
		Script        string
	}

	type ContractTeleportDataInfo struct {
		Id                   uint16
		TeleportAddress      string
		MinterAddress        string
		BitcoinClientAddress string
		CoordinatorAddress   string
		InspectorAddress     string
		ConfiguratorAddress  string
		TweakedPubkey        string
		InternalKey          string
		NextSVB              uint16
		BaseSVB              uint16
		PegoutChainCounter   uint64
		LastPegoutTxID       *chainhash.Hash
		CsvLock              uint32
		Limits               teleportcontract.Limits
		TotalServiceFee      int32
		Enabled              bool
		PeginsCount          int32
		UTXOset              *[]ContractTeleportUTXO
	}

	convertDepositsFn := func(data map[uint64]teleportcontract.DepositData) int32 {
		return int32(len(data))
	}

	convertUTXOSetFn := func(utxoSet map[string]teleportcontract.UTXOData) *[]ContractTeleportUTXO {
		contractTeleportUTXOData := []ContractTeleportUTXO{}

		for address, utxo := range utxoSet {
			cutxo := ContractTeleportUTXO{
				Address:       address,
				Amount:        utxo.Amount,
				Index:         utxo.Index,
				TapMerkleRoot: utxo.TapMerkleRoot,
				MintAddress:   utils.AddrToRawString(utxo.MintAddress),
				Script:        utxo.Script,
			}

			contractTeleportUTXOData = append(contractTeleportUTXOData, cutxo)
		}

		return &contractTeleportUTXOData
	}

	contractTeleportStorage, err := manager.dataSourceDB.TeleportContractStorage()
	if err != nil {
		return "", err
	}

	contractTeleportDataInfo := ContractTeleportDataInfo{
		Id:                   contractTeleportStorage.Id,
		TeleportAddress:      utils.AddrToRawString(manager.contractAddrs["teleport"]),
		MinterAddress:        utils.AddrToRawString(contractTeleportStorage.MinterAddress),
		BitcoinClientAddress: utils.AddrToRawString(contractTeleportStorage.BitcoinClientAddress),
		CoordinatorAddress:   utils.AddrToRawString(contractTeleportStorage.CoordinatorAddress),
		InspectorAddress:     utils.AddrToRawString(contractTeleportStorage.InspectorAddress),
		ConfiguratorAddress:  utils.AddrToRawString(contractTeleportStorage.ConfiguratorAddress),
		TweakedPubkey:        contractTeleportStorage.TweakedPubkey,
		InternalKey:          contractTeleportStorage.InternalKey,
		NextSVB:              contractTeleportStorage.NextSVB,
		BaseSVB:              contractTeleportStorage.BaseSVB,
		PegoutChainCounter:   contractTeleportStorage.PegoutChainCounter,
		LastPegoutTxID:       contractTeleportStorage.LastPegoutTxID,
		CsvLock:              contractTeleportStorage.CsvLock,
		Limits:               contractTeleportStorage.Limits,
		TotalServiceFee:      contractTeleportStorage.TotalServiceFee,
		Enabled:              contractTeleportStorage.Enabled,
		PeginsCount:          convertDepositsFn(contractTeleportStorage.Deposits),
		UTXOset:              convertUTXOSetFn(contractTeleportStorage.UTXOset),
	}

	contractBitcoinClient, err := manager.dataSourceDB.BitcoinClientContractStorage()
	if err != nil {
		return "", err
	}

	bitcoinNetworkInfo, err := manager.dataSourceDB.BitcoinNetworkInfoStorage()
	if err != nil {
		return "", err
	}

	type Result struct {
		ContractBitcoinClient *data_models.BitcoinClientContractStorage
		BitcoinNetworkInfo    *data_models.BitcoinNetworkInfo
		ContractTeleport      *ContractTeleportDataInfo
	}

	result := Result{
		ContractBitcoinClient: contractBitcoinClient,
		BitcoinNetworkInfo:    bitcoinNetworkInfo,
		ContractTeleport:      &contractTeleportDataInfo,
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

func (manager *MetricsManager) PlotMintedJson() (string, error) {
	rows, err := manager.db.Query(
		`SELECT COALESCE(jsonb_agg(result), '[]') AS data FROM (
			WITH data_by_days AS (
				SELECT
					DATE_TRUNC('day', created_at)::date AS day,
					SUM(amount::int8) AS minted,
					COUNT(1) AS count
				FROM mints
				WHERE status = 'SUCCESS'
				GROUP BY DATE_TRUNC('day', created_at)
			) SELECT day, TO_CHAR(minted::numeric(24,8) / 100000000::numeric(24,8), 'FM999999990.00000000') AS minted, count FROM data_by_days ORDER BY day ASC
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

func (manager *MetricsManager) PlotBurnedJson() (string, error) {
	rows, err := manager.db.Query(
		`SELECT COALESCE(jsonb_agg(result), '[]') AS data FROM (
			WITH data_by_days AS (
				SELECT
					DATE_TRUNC('day', tt.created_at)::date AS day,
					SUM(b.amount::int8) AS burned,
					COUNT(1) AS count
				FROM burns AS b 
				JOIN ton_txes AS tt ON tt.id = b.ton_tx_burn
				JOIN pegouts AS p ON p.id = b.pegout_burn 
				WHERE p.status = 'CONFIRMED' AND b.sender_addr != '0:'
				GROUP BY DATE_TRUNC('day', tt.created_at)
  		) SELECT day, TO_CHAR(burned::numeric(24,8) / 100000000::numeric(24,8), 'FM999999990.00000000') AS burned, count FROM data_by_days ORDER BY day ASC
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

func (manager *MetricsManager) PlotTotalSupplyJson() (string, error) {
	rows, err := manager.db.Query(
		`SELECT COALESCE(jsonb_agg(result), '[]') AS data FROM (
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
				WHERE p.status = 'CONFIRMED' AND b.sender_addr != '0:'
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
				SUM(daily_sum::numeric(24,8) / 100000000::numeric(24,8)) OVER (ORDER BY day) AS cumulative_total
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

func (manager *MetricsManager) PlotsSummaryJson() (string, error) {
	rows, err := manager.db.Query(
		`SELECT jsonb_build_object(
				'mints_count', (
						SELECT COUNT(1) AS row_count FROM mints WHERE status = 'SUCCESS'
				),
				'burns_count', (
						SELECT COUNT(1) AS row_count FROM burns AS b INNER JOIN pegouts AS p ON b.pegout_burn = p.id AND p.status = 'CONFIRMED' AND b.sender_addr != '0:'
				),
				'total_minted', (
						SELECT COALESCE(SUM(amount::int8)::numeric(24,8) / 100000000::numeric(24,8), 0) AS total_minted FROM mints WHERE status = 'SUCCESS'
				),
				'total_burned', (
						SELECT COALESCE(SUM(b.amount::int8)::numeric(24,8) / 100000000::numeric(24,8), 0) AS total_burned FROM burns AS b JOIN pegouts AS p ON p.id = b.pegout_burn WHERE p.status = 'CONFIRMED' AND b.sender_addr != '0:'
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

func (manager *MetricsManager) DkgStatusJson(ctx context.Context) (string, error) {
	type OriginalData struct {
		Dkg            *coordinator.DKG
		LastRestartDkg *coordinator.DKG
		PrevDkg        *coordinator.DKG
	}

	type DkgInfo struct {
		State              string
		VSetSize           int
		ValidatorsCountMax int

		ValidatorsCountInDkg    int
		ValidatorsCountNotInDkg int
		ValidatorsCountEvicted  int

		ValidatorsIdxInDkg    map[int]string
		ValidatorsIdxNotInDkg map[int]string
		ValidatorsIdxEvicted  map[int]string
	}

	type DkgStatus struct {
		StandaloneMode bool
		DkgInfo        DkgInfo
		PrevDkgInfo    DkgInfo
		Original       OriginalData
	}

	dkg, err := manager.dataSourceDB.Dkg()
	if err != nil {
		return "", err
	}

	lastRestartDkg, err := manager.dataSourceDB.DkgBeforeRestart(dkg.Until)
	if err != nil {
		return "", err
	}

	prevDkg, err := manager.dataSourceDB.PrevDkg()
	if err != nil {
		return "", err
	}

	coordinatorContractData, err := manager.dataSourceDB.CoordinatorContractStorage()
	if err != nil {
		return "", err
	}

	var status DkgStatus
	status.StandaloneMode = coordinatorContractData.StandaloneMode
	status.Original.Dkg = dkg
	status.Original.LastRestartDkg = lastRestartDkg
	status.Original.PrevDkg = prevDkg

	sumarizeDkgInfo := func(dkg *coordinator.DKG) (*DkgInfo, error) {
		var info DkgInfo

		info.State = dkg.State.String()
		info.VSetSize = len(dkg.VSet)

		if dkg.R1 != nil {
			info.ValidatorsCountInDkg = int(dkg.R1.Count)
		} else {
			info.ValidatorsCountInDkg = 0
		}

		if !coordinatorContractData.StandaloneMode {
			maxValidators, err := manager.globalRuntimeConfig.TonMaxMainValidators(ctx)
			if err != nil {
				return nil, err
			}

			info.ValidatorsCountMax = maxValidators
		} else {
			info.ValidatorsCountMax = info.VSetSize
		}

		info.ValidatorsCountNotInDkg = info.ValidatorsCountMax - info.ValidatorsCountInDkg
		info.ValidatorsIdxInDkg = make(map[int]string)
		info.ValidatorsIdxNotInDkg = make(map[int]string)
		info.ValidatorsIdxEvicted = make(map[int]string)

		count := min(info.VSetSize, info.ValidatorsCountMax)

		for i := 0; i < count; i++ {
			pubkey := dkg.VSet[uint16(i)]
			pubkeyBase64 := base64.StdEncoding.EncodeToString(pubkey)

			if dkg.VSetMask.Bit(i) == 1 {
				if dkg.R1.Mask.Bit(i) == 1 {
					info.ValidatorsIdxInDkg[i] = pubkeyBase64
				} else {
					info.ValidatorsIdxNotInDkg[i] = pubkeyBase64
				}
			} else {
				info.ValidatorsIdxEvicted[i] = pubkeyBase64
			}
		}

		info.ValidatorsCountEvicted = len(info.ValidatorsIdxEvicted)

		return &info, nil
	}

	dkgInfo, err := sumarizeDkgInfo(status.Original.Dkg)
	if err != nil {
		return "", err
	}
	status.DkgInfo = *dkgInfo

	prevDkgInfo, err := sumarizeDkgInfo(status.Original.PrevDkg)
	if err != nil {
		return "", err
	}
	status.PrevDkgInfo = *prevDkgInfo

	jsonData, err := json.Marshal(status)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

func (manager *MetricsManager) ContractBalanceJson(name string) (string, error) {
	rows, err := manager.db.Query(
		`SELECT COALESCE(jsonb_agg(result), '[]') AS data FROM (
			SELECT 
			  create_at AS ts,
				TO_CHAR(value::numeric(24,8) / 1000000000::numeric(24,8), 'FM999999990.00000000') AS balance
			FROM metrics_balances WHERE name = $1 ORDER BY id ASC
	  ) AS result`,
		name,
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

func (manager *MetricsManager) SystemInfoJson() (string, error) {
	var systemInfo MetricsSystemInfo
	return systemInfo.SystemInfoJson(manager.dataSourceDB, manager.alertManager, manager.contractAddrs)
}
