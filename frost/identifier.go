package frost

import (
	"encoding/hex"
)

type Identifier [32]byte

func DecodeIdentifier(s string) (*Identifier, error) {
	var identifier Identifier
	id, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	copy(identifier[:], id)
	return &identifier, nil
}

func (id Identifier) ToBytes() []byte {
	return id[:]
}
