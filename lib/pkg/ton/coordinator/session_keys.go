package coordinator

import (
	"errors"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/parseddict"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type SessionPubKeys map[uint16][]byte

type SessionKeys struct {
	PubKeys SessionPubKeys
}

func NewSessionPubKeys(dict *cell.Dictionary) (SessionPubKeys, error) {
	result, err := parseddict.ParseDict(
		dict,
		func(keySlice *cell.Slice, _ uint) uint16 {
			return uint16(keySlice.MustLoadUInt(16))
		},
		func(valueSlice *cell.Slice) ([]byte, error) {
			tag, err := valueSlice.LoadUInt(8)
			if err != nil {
				return nil, err
			}
			if tag != 0x53 {
				return nil, errors.New("invalid session key descr tag")
			}
			pubKeyTag, err := valueSlice.LoadUInt(32)
			if err != nil {
				return nil, err
			}
			if pubKeyTag != Ed25519PubkeyTag {
				return nil, errors.New("invalid public key tag")
			}

			return valueSlice.LoadSlice(256)
		},
	)

	return *result, err
}

func LoadSessionKeys(slice *cell.Slice) (*SessionKeys, error) {
	cs := slice.MustLoadRef()

	sessionPubKeysDict := cs.MustLoadDict(16)
	sessionPubKeys, err := NewSessionPubKeys(sessionPubKeysDict)
	if err != nil {
		return nil, err
	}

	return &SessionKeys{
		PubKeys: sessionPubKeys,
	}, nil
}
