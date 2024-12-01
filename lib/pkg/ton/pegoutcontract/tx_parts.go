package pegoutcontract

import "github.com/xssnick/tonutils-go/tvm/cell"

type TxParts struct {
	Inputs      *TxPartsInputs
	Outputs     []*TxPartsOutput
	Signatures  *TxPartsSignatures
	InternalKey []byte
}

type TxPartsOutput struct {
	Amount        uint64
	BitcoinScript []byte
}

func DecodeTxPartsOutput(outputCell *cell.Cell) (*TxPartsOutput, error) {
	outputSlice := outputCell.BeginParse()
	amount := outputSlice.MustLoadUInt(64)
	bitcoinScriptLen := outputSlice.MustLoadUInt(8)
	bitcoinScript := outputSlice.MustLoadSlice(uint(bitcoinScriptLen) * 8)
	return &TxPartsOutput{
		Amount:        amount,
		BitcoinScript: bitcoinScript,
	}, nil
}
