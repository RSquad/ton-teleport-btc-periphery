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
	dictCell *cell.Dictionary
}

func NewUnsignedPegoutsDict(dictCell *cell.Dictionary) *UnsignedPegoutsDict {
	return &UnsignedPegoutsDict{
		dictCell,
	}
}

func (c *UnsignedPegoutsDict) Get(key *big.Int) *UnsignedPegout {
	if c.dictCell == nil {
		return nil
	}
	valueSlice, err := c.dictCell.LoadValueByIntKey(key)
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
