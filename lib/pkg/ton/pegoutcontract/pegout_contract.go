package pegoutcontract

import (
	"context"
	"fmt"
	"math/big"

	tonclient "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

const (
	txPartsIndexInputsDictCell     = 0
	txPartsIndexPegoutOutputCell   = 1
	txPartsIndexChangeOutputCell   = 2
	txPartsIndexSignaturesDictCell = 3
	txPartsIndexInternalKey        = 4
)

type PegoutContract struct {
	Addr      *address.Address
	tonClient *tonclient.TonClient
	ctx       context.Context
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
	tonClient *tonclient.TonClient,
	ctx context.Context,
) *PegoutContract {
	return &PegoutContract{addr, tonClient, ctx}
}

func NewFromStateInit(
	stateInit *StateInit,
	tonClient *tonclient.TonClient,
	ctx context.Context,
) (*PegoutContract, error) {
	initDataCell := InitDataToCell(*stateInit.InitData)
	stateInitCell, err := tlb.ToCell(tlb.StateInit{Code: stateInit.Code, Data: initDataCell})
	if err != nil {
		return nil, fmt.Errorf("[PegoutContract] failed to build state init cell: %w", err)
	}
	addr := address.NewAddress(0, byte(stateInit.InitData.TeleportContractAddr.Workchain()), stateInitCell.Hash())
	return &PegoutContract{addr, tonClient, ctx}, nil
}

func (c *PegoutContract) GetTxParts(block *ton.BlockIDExt) (*TxParts, error) {
	res, err := c.tonClient.API.RunGetMethod(c.ctx, block, c.Addr, "get_tx_parts")
	if err != nil {
		return nil, fmt.Errorf("failed to get tx parts: %w", err)
	}

	inputsDictCell, err := res.Cell(txPartsIndexInputsDictCell)
	if err != nil {
		return nil, fmt.Errorf("failed to get inputs dict cell: %w", err)
	}
	inputs, err := NewTxPartsInputs(inputsDictCell.AsDict(256))
	if err != nil {
		return nil, fmt.Errorf("failed to decode inputs: %w", err)
	}

	pegoutOutputCell, err := res.Cell(txPartsIndexPegoutOutputCell)
	if err != nil {
		return nil, fmt.Errorf("failed to get pegout output cell: %w", err)
	}
	pegoutOutput, err := DecodeTxPartsOutput(pegoutOutputCell)
	if err != nil {
		return nil, fmt.Errorf("failed to decode pegout output: %w", err)
	}
	outputs := []*TxPartsOutput{pegoutOutput}

	changeOutputCell, err := res.Cell(txPartsIndexChangeOutputCell)
	if err == nil {
		changeOutput, err := DecodeTxPartsOutput(changeOutputCell)
		if err != nil {
			return nil, fmt.Errorf("failed to decode change output: %w", err)
		}
		outputs = append(outputs, changeOutput)
	}

	signaturesDictCell, err := res.Cell(txPartsIndexSignaturesDictCell)
	if err != nil {
		return nil, fmt.Errorf("failed to get signatures dict cell: %w", err)
	}
	signatures, err := NewTxPartsSignatures(signaturesDictCell.AsDict(16))
	if err != nil {
		return nil, fmt.Errorf("failed to decode signatures: %w", err)
	}

	internalKey, err := res.Int(txPartsIndexInternalKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get internal key: %w", err)
	}

	return &TxParts{
		Inputs:      inputs,
		Outputs:     outputs,
		Signatures:  signatures,
		InternalKey: internalKey.Bytes(),
	}, nil
}
