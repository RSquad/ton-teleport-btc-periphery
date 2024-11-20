package pegoutcontract

import (
	"context"
	"fmt"
	"math/big"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type PegoutContract struct {
	Addr *address.Address
	API  *ton.APIClient
	Ctx  context.Context
}

type StateInit struct {
	InitData *InitData
	Code     *cell.Cell
}

type InitData struct {
	ID                   uint32
	Amount               *big.Int
	BitcoinScript        []byte
	TeleportContractAddr *address.Address
}

func InitDataToCell(initData InitData) *cell.Cell {
	bitcoinScriptLen := len(initData.BitcoinScript)
	return cell.BeginCell().
		MustStoreUInt(0, 1).
		MustStoreUInt(uint64(initData.ID), 32).
		MustStoreUInt(initData.Amount.Uint64(), 64).
		MustStoreUInt(uint64(bitcoinScriptLen), 8).
		MustStoreSlice(initData.BitcoinScript, uint(bitcoinScriptLen)*8).
		MustStoreAddr(initData.TeleportContractAddr).
		EndCell()
}

func New(
	addr *address.Address,
	api *ton.APIClient,
	ctx context.Context,
) *PegoutContract {
	return &PegoutContract{addr, api, ctx}
}

func NewFromStateInit(stateInit *StateInit, api *ton.APIClient, ctx context.Context) (*PegoutContract, error) {
	initDataCell := InitDataToCell(*stateInit.InitData)
	stateInitCell, err := tlb.ToCell(tlb.StateInit{Code: stateInit.Code, Data: initDataCell})
	if err != nil {
		return nil, fmt.Errorf("[PegoutContract] failed to build state init cell: %w", err)
	}
	addr := address.NewAddress(0, byte(stateInit.InitData.TeleportContractAddr.Workchain()), stateInitCell.Hash())
	return &PegoutContract{addr, api, ctx}, nil
}
