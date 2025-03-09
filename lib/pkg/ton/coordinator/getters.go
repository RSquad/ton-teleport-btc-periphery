package coordinator

import (
	"context"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/parseddict"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/signer"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
	tonutils "github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

const (
	OpCodeStartDKG                    = 0xe16b89c0
	OpCodeCoordinatorRound1           = 0x0000eaea
	OpCodeCoordinatorRound2           = 0x0000bb50
	OpCodeCoordinatorRound3           = 0x00008bc6
	OpCodeCoordinatorSendCommitments  = 0x58e40000
	OpCodeCoordinatorSendSigningShare = 0x706b0000
	OpCodeCoordinatorSendSignature    = 0xd0720000
)

const DefaultDGKTTL = time.Minute

type CoordinatorContract struct {
	ton.Contract
	signer    *signer.Signer
	tonClient *tonclient.TonClient
	ctx       context.Context
	ttl       time.Duration
}

func New(
	addr *address.Address,
	tonClient *tonclient.TonClient,
	signer *signer.Signer,
	ctx context.Context,
) *CoordinatorContract {
	ttl := DefaultDGKTTL
	return &CoordinatorContract{
		ton.Contract{Addr: addr}, signer, tonClient, ctx, ttl,
	}
}

func (c *CoordinatorContract) GetDkg(block *tonutils.BlockIDExt) (*DKG, error) {
	if block == nil {
		var err error
		block, err = c.tonClient.API.CurrentMasterchainInfo(c.ctx)
		if err != nil {
			return nil, err
		}
	}

	result, err := c.tonClient.API.RunGetMethod(c.ctx, block, c.Addr, "get_dkg")
	if err != nil {
		return nil, err
	}

	dkg, err := parseDGKSlice(result.MustCell(0).BeginParse())
	if err != nil {
		return nil, err
	}

	return dkg, nil
}

func (c *CoordinatorContract) GetPrevDKG() (*DKG, error) {
	block, err := c.tonClient.API.CurrentMasterchainInfo(c.ctx)
	if err != nil {
		return nil, err
	}

	result, err := c.tonClient.API.RunGetMethod(c.ctx, block, c.Addr, "get_prev_dkg")
	if err != nil {
		return nil, err
	}

	return parseDGKSlice(result.MustCell(0).BeginParse())
}

func (c *CoordinatorContract) GetUnsignedPegouts() ([]PegoutRecord, error) {
	block, err := c.tonClient.API.CurrentMasterchainInfo(c.ctx)
	if err != nil {
		return nil, err
	}

	result, err := c.tonClient.API.RunGetMethod(c.ctx, block, c.Addr, "get_pegout_records")
	if err != nil {
		return nil, err
	}

	cell, err := result.Cell(0)
	if err != nil {
		return nil, err
	}
	dict, err := cell.BeginParse().ToDict(64)
	if err != nil {
		return nil, err
	}
	entries, err := dict.LoadAll()
	if err != nil {
		return nil, err
	}

	pegouts := make([]PegoutRecord, 0, len(entries))
	for _, kv := range entries {
		ID := kv.Key.MustLoadUInt(64)
		value := kv.Value.MustLoadRef()
		CommitmentsMask := value.MustLoadSlice(256)
		value.MustLoadUInt(16)
		commitmentsDict := value.MustLoadDict(256)
		commitmentsPtr, _ := parseddict.New(
			commitmentsDict,
			parseddict.ParseKey,
			readBuffer,
		)
		Commitments := *commitmentsPtr
		SigningSharesMask := value.MustLoadSlice(256)
		value.MustLoadUInt(16)
		signingSharesDict := value.MustLoadDict(256)
		signingSharesPtr, _ := parseddict.New(
			signingSharesDict,
			parseddict.ParseKey,
			loadSharesMap,
		)
		SigningShares := *signingSharesPtr
		PegoutAddress := value.MustLoadAddr()
		InternalKey := value.MustLoadRef().MustLoadSlice(256)

		pegouts = append(pegouts, PegoutRecord{
			ID,
			PegoutAddress,
			InternalKey,
			Commitments,
			CommitmentsMask,
			SigningShares, // map[string]*Cell
			SigningSharesMask,
		})
	}
	return pegouts, nil
}

func parseDGKSlice(dkgSlice *cell.Slice) (*DKG, error) {
	status := DKGStatus(dkgSlice.MustLoadUInt(2))

	vSetDictCell := dkgSlice.MustLoadDict(16)
	vSet, err := NewVSet(vSetDictCell)
	if err != nil {
		return nil, err
	}

	maxSigners := dkgSlice.MustLoadUInt(16)

	r1State, err := loadRoundMaskAndCount(dkgSlice)
	if err != nil {
		return nil, err
	}
	r1PkgDictCell := dkgSlice.MustLoadDict(256)
	r1, err := NewDKGR1(r1PkgDictCell, r1State)
	if err != nil {
		return nil, err
	}

	r2State, err := loadRoundMaskAndCount(dkgSlice)
	if err != nil {
		return nil, err
	}
	r2PkgDictCell := dkgSlice.MustLoadDict(256)
	r2, err := NewDKGR2(r2PkgDictCell, r2State)
	if err != nil {
		return nil, err
	}

	r3, err := LoadDKGR3(dkgSlice.MustLoadRef())
	if err != nil {
		return nil, err
	}

	return &DKG{
		Status:     status,
		VSet:       vSet,
		MaxSigners: maxSigners,
		R1:         r1,
		R2:         r2,
		R3:         r3,
	}, nil
}

func loadRoundMaskAndCount(dkgSlice *cell.Slice) (*DKGRoundState, error) {
	mask, err := dkgSlice.LoadBigUInt(256)
	if err != nil {
		return nil, err
	}
	count, err := dkgSlice.LoadUInt(16)
	if err != nil {
		return nil, err
	}

	return &DKGRoundState{mask, count}, nil
}

// Helpers

func readBuffer(value *cell.Slice) ([]byte, error) {
	return utils.WriteSlicesToBuffer(value.MustLoadRef()), nil
}

func loadSharesMap(value *cell.Slice) (map[string][]byte, error) {
	dict, _ := value.MustLoadRef().ToDict(64)
	sharesMap, err := parseddict.New(dict, parseddict.ParseKey, func(s *cell.Slice) ([]byte, error) {
		return utils.WriteSlicesToBuffer(s), nil
	})
	return *sharesMap, err
}
