package dict

import (
	"encoding/hex"
	"math/big"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

// PegoutRecord represents a pegout transaction record
type PegoutRecord struct {
	ID                uint64
	PegoutAddress     address.Address
	InternalKey       []byte
	Commitments       map[string][]byte
	CommitmentsMask   []byte
	SigningShares     map[string]map[string][]byte
	SigningSharesMask []byte
}

// HasCommitment checks if a commitment exists for the given identifier
func (p *PegoutRecord) HasCommitment(identifier []byte) bool {
	_, exists := p.Commitments[hex.EncodeToString(identifier)]
	return exists
}

// CommitmentsCount returns the number of commitments
func (p *PegoutRecord) CommitmentsCount() int {
	return len(p.Commitments)
}

// HasSigningShare checks if a signing share exists for the given identifier
func (p *PegoutRecord) HasSigningShare(identifier []byte) bool {
	_, exists := p.SigningShares[string(identifier)]
	return exists
}

// SigningSharesCount returns the number of signing shares
func (p *PegoutRecord) SigningSharesCount() int {
	return len(p.SigningShares)
}

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
