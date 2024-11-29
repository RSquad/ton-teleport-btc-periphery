package dict

import "github.com/xssnick/tonutils-go/tvm/cell"

type (
	SigningSharesKey   [32]byte
	SigningSharesValue *cell.Cell
)

type SigningSharesDict struct {
	Dict[SigningSharesKey, SigningSharesValue]
}

func (b *SigningSharesDict) NewDict(cellDictionary *cell.Dictionary) *Dict[SigningSharesKey, SigningSharesValue] {
	dict := &Dict[SigningSharesKey, SigningSharesValue]{
		parseKey:       b.parseKey,
		parseValue:     b.parseValue,
		cellDictionary: cellDictionary,
	}
	return &Dict[SigningSharesKey, SigningSharesValue]{dictionary: dict.Parse()}
}

func (b *SigningSharesDict) parseKey(key *cell.Slice) SigningSharesKey {
	k := key.MustLoadSlice(256)

	return SigningSharesKey(k)
}

func (b *SigningSharesDict) parseValue(value *cell.Slice) SigningSharesValue {
	return value.MustToCell()
}
