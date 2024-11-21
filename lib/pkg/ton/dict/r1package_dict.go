package dict

import "github.com/xssnick/tonutils-go/tvm/cell"

type (
	R1PackageKey   [32]byte
	R1PackageValue [][]byte
)

type R1PackageDict struct {
	Dict[R1PackageKey, R1PackageValue]
}

func (b *R1PackageDict) NewDict(cellDictionary *cell.Dictionary) *Dict[R1PackageKey, R1PackageValue] {
	dict := &Dict[R1PackageKey, R1PackageValue]{
		parseKey:       b.parseKey,
		parseValue:     b.parseValue,
		cellDictionary: cellDictionary,
	}
	return dict
}

func (b *R1PackageDict) parseKey(key *cell.Slice) R1PackageKey {
	k := key.MustLoadSlice(256)

	return R1PackageKey(k)
}

func (b *R1PackageDict) parseValue(value *cell.Slice) R1PackageValue {
	return writeCellsToBuffer(value.MustLoadRef())
}
