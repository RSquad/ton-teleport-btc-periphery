package dict

import "github.com/xssnick/tonutils-go/tvm/cell"

const Ed25519PubkeyTag = 0x8e81278a

type VSetDict struct {
	Dict[uint64, []byte]
}

func (b *VSetDict) NewDict(cellDictionary *cell.Dictionary) *Dict[uint64, []byte] {
	dict := &Dict[uint64, []byte]{
		parseKey:       b.parseKey,
		parseValue:     b.parseValue,
		cellDictionary: cellDictionary,
	}
	return &Dict[uint64, []byte]{dictionary: dict.Parse()}
}

func (b VSetDict) parseKey(key *cell.Slice) uint64 {
	return key.MustLoadUInt(16)
}

func (b VSetDict) parseValue(value *cell.Slice) []byte {
	tag := value.MustLoadUInt(8)
	if (tag &^ 0x20) != 0x53 {
		panic("Invalid Validator Descr tag")
	}
	pubKeyTag := value.MustLoadUInt(32)
	if pubKeyTag != Ed25519PubkeyTag {
		panic("Invalid PublicKey tag")
	}
	return value.MustLoadSlice(256)
}
