package credscontract

import (
	"context"
	"fmt"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	tonutils "github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/parseddict"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
)

type CredsContract struct {
	ton.Contract
	TonClient *tonclient.TonClient
}

type Storage struct {
	Owner *address.Address
	Creds map[string][]byte
}

type StateInit struct {
	InitData *InitData
	Code     *cell.Cell
}

type InitData struct {
	Owner *address.Address
}

var CODE_HEX = "b5ee9c72410107010081000114ff00f4a413f4bcf2c80b010201620206035cd020c700915be001d0d3030171b0915be0fa403001d31fdb3c018210ce2eda5eba8f0501db3cdb3ce05b840ff2f0030405001ced44d0fa4001f861f40401f862d1002af841c705f2e06ef84201d3ffd4d1028307f417f8620018f842c8f841cf16f400c9ed540009a0a75bda89c8425675"

func GetCode() *cell.Cell {
	code, err := cell.FromBOC(utils.MustHexToBytes(CODE_HEX, len(CODE_HEX)/2))
	if err != nil {
		return nil
	}
	return code
}

func InitDataToCell(initData InitData) *cell.Cell {
	return cell.BeginCell().
		MustStoreAddr(initData.Owner).
		MustStoreDict(nil).
		EndCell()
}

func NewFromStateInit(
	tonClient *tonclient.TonClient,
	stateInit *StateInit,
) (*CredsContract, error) {
	initDataCell := InitDataToCell(*stateInit.InitData)
	stateInitCell, err := tlb.ToCell(tlb.StateInit{Code: stateInit.Code, Data: initDataCell})
	if err != nil {
		return nil, fmt.Errorf("failed to build creds contract state init cell: %w", err)
	}
	addr := address.NewAddress(0, 0, stateInitCell.Hash())
	return &CredsContract{ton.Contract{Addr: addr}, tonClient}, nil
}

func (c *CredsContract) GetStorage(ctx context.Context, block *tonutils.BlockIDExt) (Storage, error) {
	if block == nil {
		var err error
		block, err = c.TonClient.API.CurrentMasterchainInfo(ctx)
		if err != nil {
			return Storage{}, err
		}
	}

	result, err := c.TonClient.API.RunGetMethod(ctx, block, c.Addr, "get_state")
	if err != nil {
		return Storage{}, err
	}

	resultCell, err := result.Cell(0)
	if err != nil {
		return Storage{}, err
	}

	resultSlice := resultCell.BeginParse()
	owner := resultSlice.MustLoadAddr()
	credsDict := resultSlice.MustLoadDict(256)
	creds, err := parseddict.ParseDict(
		credsDict,
		parseddict.ParseKeyStr,
		func(s *cell.Slice) ([]byte, error) {
			return utils.WriteSlicesToBuffer(s), nil
		},
	)
	if err != nil {
		return Storage{}, err
	}

	return Storage{
		Owner: owner,
		Creds: *creds,
	}, nil
}
