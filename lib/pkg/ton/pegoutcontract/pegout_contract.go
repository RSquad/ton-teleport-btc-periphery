package pegoutcontract

import (
	"context"
	"fmt"

	tonclient "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/ton"
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

func New(
	addr *address.Address,
	tonClient *tonclient.TonClient,
	ctx context.Context,
) *PegoutContract {
	return &PegoutContract{addr, tonClient, ctx}
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

func (c *PegoutContract) GetSigningHashes(block *ton.BlockIDExt) ([][]byte, error) {
	result, err := c.tonClient.API.RunGetMethod(c.ctx, block, c.Addr, "get_signing_hashes")
	if err != nil {
		return nil, fmt.Errorf("failed to get tx parts: %w", err)
	}
	cell, err := result.Cell(0)
	if err != nil {
		return nil, err
	}

	dict, err := cell.BeginParse().ToDict(16)
	if err != nil {
		return nil, err
	}

	entries, err := dict.LoadAll()
	if err != nil {
		return nil, err
	}
	hashes := make([][]byte, 0, len(entries))
	for _, kv := range entries {
		hashes = append(hashes, utils.WriteSlicesToBuffer(kv.Value))
	}

	return hashes, nil
}
