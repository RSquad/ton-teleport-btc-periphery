package data_sources

import (
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
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

func (dataSource *DataSourceDB) CoordinatorContractStorageJson() (map[string]interface{}, error) {
	return dataSource.SelectToObject("SELECT payload FROM metrics_data WHERE type_id = 5 ORDER BY id DESC LIMIT 1")
}

func (dataSource *DataSourceDB) CoordinatorContractStorage() (*coordinator.Storage, error) {
	json, err := dataSource.CoordinatorContractStorageJson()
	if err != nil {
		return nil, err
	}

	data, err := data_models.DeserializeCoordinatorContractStorage(json)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (dataSource *DataSourceDB) DkgJson() (map[string]interface{}, error) {
	return dataSource.SelectToObject("SELECT payload FROM metrics_data WHERE type_id = 0 ORDER BY id DESC LIMIT 1")
}

func (dataSource *DataSourceDB) Dkg() (*coordinator.DKG, error) {
	json, err := dataSource.DkgJson()
	if err != nil {
		return nil, err
	}

	data, err := data_models.DeserializeDkg(json)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (dataSource *DataSourceDB) PrevDkgJson() (map[string]interface{}, error) {
	return dataSource.SelectToObject("SELECT payload FROM metrics_data WHERE type_id = 1 ORDER BY id DESC LIMIT 1")
}

func (dataSource *DataSourceDB) PrevDkg() (*coordinator.DKG, error) {
	json, err := dataSource.PrevDkgJson()
	if err != nil {
		return nil, err
	}

	data, err := data_models.DeserializeDkg(json)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (dataSource *DataSourceDB) PegoutJson(address *address.Address) (map[string]interface{}, error) {
	return dataSource.SelectToObject(
		"SELECT row_to_json(t) FROM (SELECT * FROM public.pegouts WHERE addr=$1) t",
		address.StringRaw(),
	)
}

func (dataSource *DataSourceDB) PegoutDbRow(address *address.Address) (*data_models.PegoutDbRow, error) {
	pegoutJson, err := dataSource.PegoutJson(address)
	if err != nil {
		return nil, err
	}

	pegout, err := data_models.DeserializePegoutDbRow(pegoutJson)
	if err != nil {
		return nil, err
	}

	return pegout, nil
}

func (dataSource *DataSourceDB) SelectToObject(sql string, args ...interface{}) (map[string]interface{}, error) {
	rows, err := dataSource.db.Query(sql, args...)
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
