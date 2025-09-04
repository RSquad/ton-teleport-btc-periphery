package data_models

import (
	"encoding/json"

	"github.com/xssnick/tonutils-go/address"
)

type ContractBalance struct {
	Name    string
	Addr    *address.Address
	Balance uint64
}

type ContractBalances struct {
	Balances []*ContractBalance
}

func DeserializeContractBalancesDB(jsonData []byte) (*ContractBalances, error) {
	var balances ContractBalances
	err := json.Unmarshal(jsonData, &balances)
	if err != nil {
		return nil, err
	}

	return &balances, nil
}

func SerializeContractBalancesDB(balances *ContractBalances) ([]byte, error) {
	return json.Marshal(balances)
}
