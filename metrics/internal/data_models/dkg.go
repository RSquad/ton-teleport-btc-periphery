package data_models

import (
	"encoding/json"
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
)

func DeserializeDkg(jsonData []byte) (*coordinator.DKG, error) {
	var dkg coordinator.DKG
	err := json.Unmarshal(jsonData, &dkg)
	if err != nil {
		return nil, fmt.Errorf("failed to call `DeserializeDkg`, json '%s': %w", string(jsonData), err)
	}

	return &dkg, nil
}
