package dict

import (
	"math/big"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type (
	UnsignedPegout struct {
		internalKey        []byte
		pegoutContractAddr *address.Address
		Commitments        *cell.Dictionary
		commitmentsMask    []byte
		SigningShares      *cell.Dictionary
		signingSharesMask  []byte
	}
)

type UnsignedPegoutsDict struct {
	dict *cell.Dictionary
}

func NewUnsignedPegoutsDict(dict *cell.Dictionary) *UnsignedPegoutsDict {
	return &UnsignedPegoutsDict{
		dict,
	}
}

func (c *UnsignedPegoutsDict) Get(key *big.Int) *UnsignedPegout {
	if c.dict == nil {
		return nil
	}
	valueSlice, err := c.dict.LoadValueByIntKey(key)
	if err != nil {
		return nil
	}

	commitmentsMask := valueSlice.MustLoadSlice(256)
	valueSlice.MustLoadUInt(16)
	commitmentsDict := valueSlice.MustLoadDict(256)
	signingSharesMask := valueSlice.MustLoadSlice(256)
	valueSlice.MustLoadUInt(16)
	signingSharesDict := valueSlice.MustLoadDict(256)
	pegoutContractAddr := valueSlice.MustLoadAddr()
	internalKey := valueSlice.MustLoadRef().MustLoadSlice(256)

	return &UnsignedPegout{
		internalKey:        internalKey,
		pegoutContractAddr: pegoutContractAddr,
		Commitments:        commitmentsDict,
		SigningShares:      signingSharesDict,
		commitmentsMask:    commitmentsMask,
		signingSharesMask:  signingSharesMask,
	}
}
