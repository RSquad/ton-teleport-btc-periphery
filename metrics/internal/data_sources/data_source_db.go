package data_sources

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/fetchers"
	"github.com/xssnick/tonutils-go/address"
)

type DataSourceDB struct {
	db *sql.DB
}

func NewDataSourceDB(db *sql.DB) *DataSourceDB {
	return &DataSourceDB{
		db: db,
	}
}

func (dataSource *DataSourceDB) CoordinatorContractStorageJson() ([]byte, error) {
	return dataSource.selectAsJsonObj(
		"SELECT payload FROM metrics_data WHERE type_id = $1 ORDER BY id DESC LIMIT 1",
		fetchers.PayloadTypeContractCoordinator,
	)
}

func (dataSource *DataSourceDB) CoordinatorContractStorage() (*coordinator.Storage, error) {
	jsonData, err := dataSource.CoordinatorContractStorageJson()
	if err != nil {
		return nil, err
	}

	data, err := data_models.DeserializeCoordinatorContractStorage(jsonData)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (dataSource *DataSourceDB) TeleportContractStorageJson() ([]byte, error) {
	return dataSource.selectAsJsonObj(
		"SELECT payload FROM metrics_data WHERE type_id = $1 ORDER BY id DESC LIMIT 1",
		fetchers.PayloadTypeContractTeleport,
	)
}

func (dataSource *DataSourceDB) TeleportContractStorage() (*teleportcontract.Storage, error) {
	jsonData, err := dataSource.TeleportContractStorageJson()
	if err != nil {
		return nil, err
	}

	data, err := data_models.DeserializeTeleportContractStorage(jsonData)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (dataSource *DataSourceDB) BitcoinClientContractStorageJson() ([]byte, error) {
	return dataSource.selectAsJsonObj(
		"SELECT payload FROM metrics_data WHERE type_id = $1 ORDER BY id DESC LIMIT 1",
		fetchers.PayloadTypeContractBitcoinClient,
	)
}

func (dataSource *DataSourceDB) BitcoinClientContractStorage() (*data_models.BitcoinClientContractStorage, error) {
	jsonData, err := dataSource.BitcoinClientContractStorageJson()
	if err != nil {
		return nil, err
	}

	data, err := data_models.DeserializeBitcoinContractStorageDB(jsonData)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (dataSource *DataSourceDB) BitcoinNetworkInfoJson() ([]byte, error) {
	return dataSource.selectAsJsonObj(
		"SELECT payload FROM metrics_data WHERE type_id = $1 ORDER BY id DESC LIMIT 1",
		fetchers.PayloadTypeBitcoinNetwork,
	)
}

func (dataSource *DataSourceDB) BitcoinNetworkInfoStorage() (*data_models.BitcoinNetworkInfo, error) {
	jsonData, err := dataSource.BitcoinNetworkInfoJson()
	if err != nil {
		return nil, err
	}

	data, err := data_models.DeserializeBitcoinNetworkInfoDB(jsonData)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (dataSource *DataSourceDB) DkgJson() ([]byte, error) {
	return dataSource.selectAsJsonObj(
		"SELECT payload FROM metrics_data WHERE type_id = $1 ORDER BY id DESC LIMIT 1",
		fetchers.PayloadTypeDKG,
	)
}

func (dataSource *DataSourceDB) Dkg() (*coordinator.DKG, error) {
	jsonData, err := dataSource.DkgJson()
	if err != nil {
		return nil, err
	}

	data, err := data_models.DeserializeDkg(jsonData)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (dataSource *DataSourceDB) PrevDkgJson() ([]byte, error) {
	return dataSource.selectAsJsonObj(
		"SELECT payload FROM metrics_data WHERE type_id = $1 ORDER BY id DESC LIMIT 1",
		fetchers.PayloadTypePrevDKG,
	)
}

func (dataSource *DataSourceDB) PrevDkg() (*coordinator.DKG, error) {
	jsonData, err := dataSource.PrevDkgJson()
	if err != nil {
		return nil, err
	}

	data, err := data_models.DeserializeDkg(jsonData)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (dataSource *DataSourceDB) DkgUntilJson(until time.Time) ([]byte, error) {
	return dataSource.selectAsJsonObj(
		"SELECT payload FROM metrics_data WHERE type_id = $1 AND dkg_until_ts = $2 ORDER BY id DESC LIMIT 1",
		fetchers.PayloadTypeDKG,
		time.Unix(until.Unix(), 0).UTC(),
	)
}

func (dataSource *DataSourceDB) DkgUntil(until time.Time) (*coordinator.DKG, error) {
	jsonData, err := dataSource.DkgUntilJson(until)
	if err != nil {
		return nil, err
	}

	if len(jsonData) == 0 {
		return nil, nil
	}

	data, err := data_models.DeserializeDkg(jsonData)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (dataSource *DataSourceDB) DkgBeforeRestartJson(currentDkgUntil time.Time) ([]byte, error) {
	return dataSource.selectAsJsonObj(
		"SELECT payload FROM metrics_data WHERE type_id = $1 AND dkg_until_ts < $2 ORDER BY id DESC LIMIT 1",
		fetchers.PayloadTypeDKG,
		time.Unix(currentDkgUntil.Unix(), 0).UTC(),
	)
}

func (dataSource *DataSourceDB) DkgBeforeRestart(currentDkgUntil time.Time) (*coordinator.DKG, error) {
	jsonData, err := dataSource.DkgBeforeRestartJson(currentDkgUntil)
	if err != nil {
		return nil, err
	}

	if len(jsonData) == 0 {
		return nil, nil
	}

	data, err := data_models.DeserializeDkg(jsonData)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (dataSource *DataSourceDB) PegoutJson(address *address.Address) ([]byte, error) {
	return dataSource.selectAsJsonObj(
		"SELECT row_to_json(t) FROM (SELECT * FROM public.pegouts WHERE addr=$1) t",
		address.StringRaw(),
	)
}

func (dataSource *DataSourceDB) Pegout(address *address.Address) (*data_models.Pegout, error) {
	jsonData, err := dataSource.PegoutJson(address)
	if err != nil {
		return nil, err
	}

	if jsonData == nil {
		return nil, nil
	}

	pegout, err := data_models.DeserializePegoutDB(jsonData, address.StringRaw())
	if err != nil {
		return nil, err
	}

	return pegout, nil
}

func (dataSource *DataSourceDB) LastConfirmedPegoutJson() ([]byte, error) {
	return dataSource.selectAsJsonObj(
		"SELECT row_to_json(t) FROM (SELECT * FROM public.pegouts WHERE status = 'CONFIRMED' ORDER BY id DESC LIMIT 1) t",
	)
}

func (dataSource *DataSourceDB) LastConfirmedPegout() (*data_models.Pegout, error) {
	jsonData, err := dataSource.LastConfirmedPegoutJson()
	if err != nil {
		return nil, err
	}

	if jsonData == nil {
		return nil, nil
	}

	pegout, err := data_models.DeserializePegoutDB(jsonData, "")
	if err != nil {
		return nil, err
	}

	return pegout, nil
}

func (dataSource *DataSourceDB) LastSignedPegoutJson() ([]byte, error) {
	return dataSource.selectAsJsonObj(
		"SELECT row_to_json(t) FROM (SELECT * FROM public.pegouts WHERE status = 'SIGNED' ORDER BY id DESC LIMIT 1) t",
	)
}

func (dataSource *DataSourceDB) LastSignedPegout() (*data_models.Pegout, error) {
	jsonData, err := dataSource.LastSignedPegoutJson()
	if err != nil {
		return nil, err
	}

	if jsonData == nil {
		return nil, nil
	}

	pegout, err := data_models.DeserializePegoutDB(jsonData, "")
	if err != nil {
		return nil, err
	}

	return pegout, nil
}

func (dataSource *DataSourceDB) LastSignedPegoutsJson(limit uint) ([]byte, error) {
	return dataSource.selectAsJsonObj(
		"SELECT COALESCE(jsonb_agg(t), '[]') FROM (SELECT * FROM public.pegouts WHERE status = 'SIGNED' ORDER BY id DESC LIMIT $1) t",
		limit,
	)
}

func (dataSource *DataSourceDB) LastSignedPegouts(limit uint) ([]*data_models.Pegout, error) {
	jsonData, err := dataSource.LastSignedPegoutsJson(limit)
	if err != nil {
		return nil, err
	}

	pegouts, err := data_models.DeserializePegoutsDB(jsonData, "")
	if err != nil {
		return nil, err
	}

	return pegouts, nil
}

func (dataSource *DataSourceDB) ActualContractBalance(name string) (int64, error) {
	rows, err := dataSource.db.Query(
		"SELECT value FROM metrics_balances WHERE name = $1 ORDER BY id DESC LIMIT 1",
		name,
	)
	if err != nil {
		return 0, err
	}

	defer rows.Close()

	balance := int64(0)
	if rows.Next() {
		err = rows.Scan(&balance)
		if err != nil {
			return 0, err
		}
	} else {
		return -1, fmt.Errorf("balance not found: '%s'", name)
	}

	return balance, nil
}

func (dataSource *DataSourceDB) selectAsJsonObj(query string, args ...interface{}) ([]byte, error) {
	logger.Log.Debug().Str("component", "DataSourceDB").Msgf("SQL: '%s'. ARGS: %#v", query, args)

	rows, err := dataSource.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var data string
	if rows.Next() {
		err = rows.Scan(&data)
		if err != nil {
			return nil, err
		}
	}

	if len(data) == 0 {
		return nil, nil
	}

	return []byte(data), nil
}

func (dataSource *DataSourceDB) EventsLastDkgStarted() (*coordinator.DKGStartedEvent, error) {
	events, err := dataSource.selectAsTonEvents(
		"SELECT * FROM public.events_data WHERE event_id=$1 ORDER BY tx_lt DESC LIMIT 1",
		coordinator.EventIdDKGStarted,
	)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return nil, nil
	}

	event := events[0]

	dkgStartedEvent, ok := event.(*coordinator.DKGStartedEvent)
	if !ok {
		return nil, fmt.Errorf("event is not *coordinator.DKGStartedEvent, got %T", event)
	}

	return dkgStartedEvent, nil
}

func (dataSource *DataSourceDB) EventsAllFromDkgRestart(fromTxLT uint64) ([]*coordinator.DKGRestartedEvent, error) {
	events, err := dataSource.selectAsTonEvents(
		"SELECT * FROM public.events_data WHERE event_id=$1 AND tx_lt >= $2 ORDER BY tx_lt ASC",
		coordinator.EventIdDKGRestarted,
		fromTxLT,
	)
	if err != nil {
		return nil, err
	}

	dkgRestartedEvents := make([]*coordinator.DKGRestartedEvent, len(events))
	for _, event := range events {
		dkgRestartedEvent, ok := event.(*coordinator.DKGRestartedEvent)
		if !ok {
			return nil, fmt.Errorf("event is not *coordinator.DKGRestartedEvent, got %T", event)
		}

		dkgRestartedEvents = append(dkgRestartedEvents, dkgRestartedEvent)
	}

	return dkgRestartedEvents, nil
}

func (dataSource *DataSourceDB) selectAsTonEvents(query string, args ...interface{}) ([]ton.EventInterface, error) {
	logger.Log.Debug().Str("component", "DataSourceDB").Msgf("SQL: '%s'. ARGS: %#v", query, args)

	rows, err := dataSource.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ton.EventInterface

	for rows.Next() {
		var id int64
		var createAt time.Time
		var eventId int64
		var addrStr string
		var txHash []byte
		var txLT int64
		var txUTime time.Time
		var bodyData []byte

		err := rows.Scan(
			&id,
			&createAt,
			&eventId,
			&addrStr,
			&txHash,
			&txLT,
			&txUTime,
			&bodyData,
		)
		if err != nil {
			return nil, err
		}

		addr, err := address.ParseRawAddr(addrStr)
		if err != nil {
			return nil, err
		}

		event, err := (&coordinator.EventParser{}).ParseJson(bodyData, uint64(eventId))
		if err != nil {
			return nil, err
		}

		evRaw := &ton.RawEvent{
			Addr:    addr,
			TxHash:  txHash,
			TxLT:    uint64(txLT),
			TxUtime: txUTime,
			Body:    nil,
		}

		event.SetRaw(evRaw)

		result = append(result, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
