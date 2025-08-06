package pegoutcontract

import (
	"context"
	"fmt"
	"math/big"

	tonclient "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
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

type PegoutContractStorage struct {
	Initiated          bool
	BitcoinTxId        *big.Int
	TeleportAddress    *address.Address
	Amount             *big.Int
	OutputScript       []byte
	Signatures         *cell.Dictionary
	UtxoSet            *cell.Dictionary
	TxFee              *big.Int
	ChangeScript       []byte
	InternalKey        *big.Int
	CoordinatorAddress *address.Address
}

func New(
	addr *address.Address,
	tonClient *tonclient.TonClient,
	ctx context.Context,
) *PegoutContract {
	return &PegoutContract{addr, tonClient, ctx}
}

func (c *PegoutContract) GetStorage(block *ton.BlockIDExt) (PegoutContractStorage, error) {
	if block == nil {
		var err error
		block, err = c.tonClient.API.CurrentMasterchainInfo(c.ctx)
		if err != nil {
			return PegoutContractStorage{}, fmt.Errorf("failed to get masterchain info: %w", err)
		}
	}
	acc, err := c.tonClient.FetchAcc(c.Addr, block)
	if err != nil {
		return PegoutContractStorage{}, fmt.Errorf("failed to fetch addcount: %w", err)
	}
	storageSlice := acc.Data.BeginParse()

	initiated, err := storageSlice.LoadBoolBit()
	if err != nil {
		return PegoutContractStorage{}, fmt.Errorf("failed to get initiated: %w", err)
	}
	bitcoinTxId, err := storageSlice.LoadBigUInt(256)
	if err != nil {
		return PegoutContractStorage{}, fmt.Errorf("failed to get bitcoin tx id: %w", err)
	}
	teleportAddress, err := storageSlice.LoadAddr()
	if err != nil {
		return PegoutContractStorage{}, fmt.Errorf("failed to get teleport address: %w", err)
	}

	uninitiatedStorage := PegoutContractStorage{
		Initiated:          initiated,
		BitcoinTxId:        bitcoinTxId,
		TeleportAddress:    teleportAddress,
		Amount:             nil,
		OutputScript:       nil,
		Signatures:         nil,
		UtxoSet:            nil,
		TxFee:              nil,
		ChangeScript:       nil,
		InternalKey:        nil,
		CoordinatorAddress: &address.Address{},
	}

	if !initiated {
		return uninitiatedStorage, nil
	}

	amount, err := storageSlice.LoadBigUInt(64)
	if err != nil {
		return uninitiatedStorage, fmt.Errorf("failed to get amount: %w", err)
	}
	txFee, err := storageSlice.LoadBigUInt(64)
	if err != nil {
		return uninitiatedStorage, err
	}
	size, err := storageSlice.LoadUInt(8)
	if err != nil {
		return uninitiatedStorage, err
	}
	outputScript, err := storageSlice.LoadSlice(uint(size))
	if err != nil {
		return uninitiatedStorage, err
	}
	signatures, err := storageSlice.LoadDict(256) // TODO: check keysize
	if err != nil {
		return uninitiatedStorage, err
	}

	utxoSet, err := storageSlice.LoadDict(256)
	if err != nil {
		return uninitiatedStorage, err
	}

	storageSlice, err = storageSlice.LoadRef()
	if err != nil {
		return uninitiatedStorage, err
	}
	size, err = storageSlice.LoadUInt(8)
	if err != nil {
		return uninitiatedStorage, err
	}
	changeScript, err := storageSlice.LoadSlice(uint(size))
	if err != nil {
		return uninitiatedStorage, err
	}
	internalKey, err := storageSlice.LoadBigUInt(256)
	if err != nil {
		return uninitiatedStorage, err
	}
	coordinatorAddress, err := storageSlice.LoadAddr()
	if err != nil {
		return uninitiatedStorage, err
	}

	return PegoutContractStorage{
		Initiated:          initiated,
		BitcoinTxId:        bitcoinTxId,
		TeleportAddress:    teleportAddress,
		Amount:             amount,
		OutputScript:       outputScript,
		Signatures:         signatures,
		UtxoSet:            utxoSet,
		TxFee:              txFee,
		ChangeScript:       changeScript,
		InternalKey:        internalKey,
		CoordinatorAddress: coordinatorAddress,
	}, nil
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

	noSignaturesYet, err := res.IsNil(txPartsIndexSignaturesDictCell)
	if err != nil {
		return nil, fmt.Errorf("failed to check if signature cell is null: %w", err)
	}
	var signatures *TxPartsSignatures
	if noSignaturesYet {
		signatures = &TxPartsSignatures{}
	} else {
		signaturesDictCell, err := res.Cell(txPartsIndexSignaturesDictCell)
		if err != nil {
			return nil, fmt.Errorf("failed to get signatures dict cell: %w", err)
		}
		signatures, err = NewTxPartsSignatures(signaturesDictCell.AsDict(16))
		if err != nil {
			return nil, fmt.Errorf("failed to decode signatures: %w", err)
		}
	}

	internalKey, err := res.Int(txPartsIndexInternalKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get internal key: %w", err)
	}

	return &TxParts{
		Inputs:      inputs,
		Outputs:     outputs,
		Signatures:  signatures,
		InternalKey: internalKey.FillBytes(make([]byte, 32)),
	}, nil
}

func (c *PegoutContract) GetSigningHashes(block *ton.BlockIDExt) ([][]byte, error) {
	result, err := c.tonClient.API.RunGetMethod(c.ctx, block, c.Addr, "get_signing_hashes")
	if err != nil {
		return nil, fmt.Errorf("failed to get signing hashes: %w", err)
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
