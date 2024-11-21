package dict_parser

import "github.com/xssnick/tonutils-go/tvm/cell"

const Ed25519PubkeyTag = 0x8e81278a

type VsetDictParser struct {
	DictParser[uint64, []byte]
}

func (b *VsetDictParser) BuildParse(dictKV []cell.DictKV) *DictParser[uint64, []byte] {
	return &DictParser[uint64, []byte]{
		ParseKey:   b.ParseKey,
		ParseValue: b.ParseValue,
		dictKV:     dictKV,
	}
}

func (b VsetDictParser) ParseKey(key *cell.Slice) (uint64, error) {
	return key.LoadUInt(16)
}

func (b VsetDictParser) ParseValue(value *cell.Slice) ([]byte, error) {
	tag, _ := value.LoadUInt(8)
	if (tag &^ 0x20) != 0x53 {
		panic("Invalid Validator Descr tag")
	}
	pubKeyTag, _ := value.LoadUInt(32)
	if pubKeyTag != Ed25519PubkeyTag {
		panic("Invalid PublicKey tag")
	}
	return value.LoadSlice(256)
}
