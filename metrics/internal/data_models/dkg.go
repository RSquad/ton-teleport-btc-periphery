package data_models

import (
	"encoding/json"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
)

func DeserializeDkg(jsonData []byte) (*coordinator.DKG, error) {
	var dkg coordinator.DKG
	err := json.Unmarshal(jsonData, &dkg)
	if err != nil {
		return nil, err
	}

	return &dkg, nil
}
