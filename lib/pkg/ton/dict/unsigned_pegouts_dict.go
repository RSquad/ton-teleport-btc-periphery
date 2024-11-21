package dict

import (
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type (
	UnsignedPegoutsKey   uint64
	UnsignedPegoutsValue struct {
		internalKey    [32]byte
		pegoutAddress  *address.Address
		commitments    map[CommitmentsKey]CommitmentsValue
		signingShares  map[SigningSharesKey]SigningSharesValue
		commitmentMask []byte
		signSharesMask []byte
	}
)

type UnsignedPegoutsDict struct {
	Dict[UnsignedPegoutsKey, UnsignedPegoutsValue]
}

func (b *UnsignedPegoutsDict) NewDict(cellDictionary *cell.Dictionary) *Dict[UnsignedPegoutsKey, UnsignedPegoutsValue] {
	dict := &Dict[UnsignedPegoutsKey, UnsignedPegoutsValue]{
		parseKey:       b.parseKey,
		parseValue:     b.parseValue,
		cellDictionary: cellDictionary,
	}
	return dict
}

func (b *UnsignedPegoutsDict) parseKey(key *cell.Slice) UnsignedPegoutsKey {
	return UnsignedPegoutsKey(key.MustLoadUInt(64))
}

func (b *UnsignedPegoutsDict) parseValue(value *cell.Slice) UnsignedPegoutsValue {
	slice := value.MustLoadRef()
	commitmentMask := slice.MustLoadSlice(32)
	slice.MustLoadUInt(16)
	commitmentsDict := CommitmentsDict{}
	commitments := commitmentsDict.NewDict(slice.MustLoadDict(16)).Get()

	signSharesMask := slice.MustLoadSlice(32)
	slice.MustLoadUInt(16)

	signingSharesDict := SigningSharesDict{}
	signingShares := signingSharesDict.NewDict(slice.MustLoadDict(16)).Get()
	pegoutAddress := slice.MustLoadAddr()
	internalKey := slice.MustLoadRef().MustLoadSlice(256)

	return UnsignedPegoutsValue{
		commitmentMask: commitmentMask,
		commitments:    commitments,
		signSharesMask: signSharesMask,
		signingShares:  signingShares,
		pegoutAddress:  pegoutAddress,
		internalKey:    [32]byte(internalKey),
	}
}
