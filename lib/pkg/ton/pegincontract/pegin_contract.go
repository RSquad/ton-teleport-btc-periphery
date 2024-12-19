package pegincontract

import (
	"context"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	tonclient "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type PeginContract struct {
	Addr      *address.Address
	tonClient *tonclient.TonClient
	ctx       context.Context
}

type StateInit struct {
	InitData *InitData
	Code     *cell.Cell
}

type InitData struct {
	BitcoinTxID          *chainhash.Hash
	TeleportContractAddr *address.Address
}

func InitDataToCell(initData InitData) *cell.Cell {
	return cell.BeginCell().
		MustStoreUInt(0, 1).
		MustStoreSlice(initData.BitcoinTxID.CloneBytes(), 256).
		MustStoreAddr(initData.TeleportContractAddr).
		EndCell()
}

func New(
	addr *address.Address,
	tonClient *tonclient.TonClient,
	ctx context.Context,
) *PeginContract {
	return &PeginContract{addr, tonClient, ctx}
}

func NewFromStateInit(
	stateInit *StateInit,
	tonClient *tonclient.TonClient,
	ctx context.Context,
) (*PeginContract, error) {
	initDataCell := InitDataToCell(*stateInit.InitData)
	stateInitCell, err := tlb.ToCell(tlb.StateInit{Code: stateInit.Code, Data: initDataCell})
	if err != nil {
		return nil, fmt.Errorf("failed to build pegin contract state init cell: %w", err)
	}
	addr := address.NewAddress(0, byte(stateInit.InitData.TeleportContractAddr.Workchain()), stateInitCell.Hash())
	return &PeginContract{addr, tonClient, ctx}, nil
}
