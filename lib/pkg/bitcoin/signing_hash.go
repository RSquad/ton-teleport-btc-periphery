package bitcoin

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/pegoutcontract"
)

type Input struct {
	Amount uint64
}

func BuildTaprootSigningHashes(
	inputs pegoutcontract.TxPartsInputs,
	outputs []*pegoutcontract.TxPartsOutput,
) ([][]byte, error) {
	tx := wire.NewMsgTx(2)

	prevOutputFetcher := make(map[wire.OutPoint]*wire.TxOut)

	for txHash, input := range inputs {
		hash, err := hex.DecodeString(txHash)
		if err != nil {
			return nil, fmt.Errorf("error decoding tx hash %s: %v", txHash, err)
		}
		prevTxHash, err := chainhash.NewHash(hash)
		if err != nil {
			return nil, fmt.Errorf("error creating hash: %v", err)
		}
		outPoint := wire.NewOutPoint(prevTxHash, input.Index)
		tx.AddTxIn(wire.NewTxIn(outPoint, nil, nil))

		prevOutputFetcher[*outPoint] = &wire.TxOut{
			PkScript: input.BitcoinScript,
			Value:    input.Amount.Int64(),
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
