package dict

import "github.com/xssnick/tonutils-go/tvm/cell"

type (
	CommitmentsKey   [32]byte
	CommitmentsValue [][]byte
)

type CommitmentsDict struct {
	Dict[CommitmentsKey, CommitmentsValue]
}

func (b *CommitmentsDict) NewDict(cellDictionary *cell.Dictionary) *Dict[CommitmentsKey, CommitmentsValue] {
	dict := &Dict[CommitmentsKey, CommitmentsValue]{
		parseKey:       b.parseKey,
		parseValue:     b.parseValue,
		cellDictionary: cellDictionary,
	}
	return dict
}

func (b *CommitmentsDict) parseKey(key *cell.Slice) CommitmentsKey {
	k := key.MustLoadSlice(256)

	return CommitmentsKey(k)
}

func (b *CommitmentsDict) parseValue(value *cell.Slice) CommitmentsValue {
	return writeCellsToBuffer(value.MustLoadRef())
}
