package data_models

import (
	"encoding/json"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

type BitcoinClientContractStorage struct {
	CandidateBlockHashes     []*chainhash.Hash
	LastConfirmedBlockHash   *chainhash.Hash
	ConfirmationsNeeded      int64
	LastConfirmedBlockHeight int64
}

func DeserializeBitcoinContractStorageDB(jsonData []byte) (*BitcoinClientContractStorage, error) {
	var storage BitcoinClientContractStorage
	err := json.Unmarshal(jsonData, &storage)
	if err != nil {
		return nil, err
	}

	return &storage, nil
}

func SerializeBitcoinContractStorageDB(storage *BitcoinClientContractStorage) ([]byte, error) {
	return json.Marshal(storage)
}
