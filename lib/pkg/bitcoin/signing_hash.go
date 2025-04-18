package bitcoin

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/pegoutcontract"
)

const TELEPORT_DEFAULT_SEQUENCE = uint32(0xFFFFFFFD)

func BuildTaprootSigningHashes(
	inputs pegoutcontract.TxPartsInputs,
	outputs []*pegoutcontract.TxPartsOutput,
) ([][]byte, error) {
	tx := wire.NewMsgTx(2)

	prevOutputFetcher := make(map[wire.OutPoint]*wire.TxOut)

	sortedInputs, err := inputs.ToSortedSlice()
	if err != nil {
		return nil, fmt.Errorf("error sorting inputs: %v", err)
	}

	for _, input := range sortedInputs {
		prevTxHash, err := chainhash.NewHash(input.TxHash)
		if err != nil {
			return nil, fmt.Errorf("error creating hash: %v", err)
		}
		outPoint := wire.NewOutPoint(prevTxHash, input.Data.Index)
		txIn := wire.NewTxIn(outPoint, nil, nil)
		txIn.Sequence = TELEPORT_DEFAULT_SEQUENCE
		tx.AddTxIn(txIn)

		prevOutputFetcher[*outPoint] = &wire.TxOut{
			PkScript: input.Data.BitcoinScript,
			Value:    input.Data.Amount.Int64(),
		}
	}

	for _, output := range outputs {
		txOut := wire.NewTxOut(int64(output.Amount), output.BitcoinScript)
		tx.AddTxOut(txOut)
	}

	fetcher := txscript.NewMultiPrevOutFetcher(prevOutputFetcher)
	cachedHashes := txscript.NewTxSigHashes(tx, fetcher)

	hashType := txscript.SigHashDefault
	inputCount := len(inputs)
	sigHashes := make([][]byte, inputCount)

	for i := 0; i < inputCount; i++ {
		sigHash, err := txscript.CalcTaprootSignatureHash(
			cachedHashes, hashType, tx, i, fetcher,
		)
		if err != nil {
			return nil, fmt.Errorf("error calculating taproot signature hash for input %d: %v", i, err)
		}
		sigHashes[i] = sigHash
	}

	return sigHashes, nil
}
