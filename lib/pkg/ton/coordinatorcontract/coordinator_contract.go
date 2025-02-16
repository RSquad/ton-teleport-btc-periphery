package coordinatorcontract

import (
	"context"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/signer"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/xssnick/tonutils-go/address"
	tonutils "github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

const (
	OpCodeStartDKG                = 0xe16b89c0
	OpCodeCoordinatorRound1       = 0x0000eaea
	OpCodeCoordinatorRound2       = 0x0000bb50
	OpCodeCoordinatorRound3       = 0x00008bc6
	OpCodePegOutTxSendCommitments = 0x58e40000
)

type CoordinatorContract struct {
	ton.Contract
	signer    *signer.Signer
	tonClient *tonclient.TonClient
	ctx       context.Context
}

func New(
	signer *signer.Signer,
	addr *address.Address,
	tonClient *tonclient.TonClient,
	ctx context.Context,
) *CoordinatorContract {
	return &CoordinatorContract{
		ton.Contract{Addr: addr}, signer, tonClient, ctx,
	}
}

func (c *CoordinatorContract) GetDkg(block *tonutils.BlockIDExt) (*DKG, error) {
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

func parseDGKSlice(dkgSlice *cell.Slice) (*DKG, error) {
	status := DKGStatus(dkgSlice.MustLoadUInt(2))

	vSetDictCell := dkgSlice.MustLoadDict(16)
	vSet, err := NewVSet(vSetDictCell)
	if err != nil {
		return nil, err
	}

	maxSigners := dkgSlice.MustLoadUInt(16)

	r1State, err := loadDKGRoundState(dkgSlice)
	if err != nil {
		return nil, err
	}
	r1PkgDictCell := dkgSlice.MustLoadDict(256)
	r1, err := NewDKGR1(r1PkgDictCell, r1State)
	if err != nil {
		return nil, err
	}

	r2State, err := loadDKGRoundState(dkgSlice)
	if err != nil {
		return nil, err
	}
	r2PkgDictCell := dkgSlice.MustLoadDict(256)
	r2, err := NewDKGR2(r2PkgDictCell, r2State)
	if err != nil {
		return nil, err
	}

	return &DKG{
		status:     status,
		vSet:       vSet,
		maxSigners: maxSigners,
		r1:         r1,
		r2:         r2,
	}, nil
}

func loadDKGRoundState(dkgSlice *cell.Slice) (*DKGRoundState, error) {
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
