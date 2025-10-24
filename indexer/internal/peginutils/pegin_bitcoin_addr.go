package peginutils

import (
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/txscript"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
)

func buildOpReturnScript(receiverAddr *address.Address) ([]byte, error) {
	return txscript.NewScriptBuilder().
		AddOp(txscript.OP_RETURN).
		AddData(bitcoin.TonAddrToBytesForTapLeaf(receiverAddr)).
		Script()
}

func buildCSVScript(recoveryKey *btcec.PublicKey, csvLock uint32) ([]byte, error) {
	return txscript.NewScriptBuilder().
		AddInt64(int64((1 << 22) | csvLock)).
		AddOp(txscript.OP_CHECKSEQUENCEVERIFY).
		AddOp(txscript.OP_DROP).
		AddData(utils.BytesPadTo(recoveryKey.X().Bytes(), 32)).
		AddOp(txscript.OP_CHECKSIG).
		Script()
}

func buildTaprootScriptTree(csvScript, opReturnScript []byte) *txscript.IndexedTapScriptTree {
	return txscript.AssembleTaprootScriptTree(
		txscript.NewBaseTapLeaf(csvScript),
		txscript.NewBaseTapLeaf(opReturnScript),
	)
}

func CalcPeginBitcoinScript(
	internalKey *btcec.PublicKey,
	recoveryKey *btcec.PublicKey,
	receiverAddr *address.Address,
	csvLock uint32,
) ([]byte, error) {
	csvScript, err := buildCSVScript(recoveryKey, csvLock)
	if err != nil {
		return nil, err
	}

	opReturnScript, err := buildOpReturnScript(receiverAddr)
	if err != nil {
		return nil, err
	}

	taprootScriptTree := buildTaprootScriptTree(csvScript, opReturnScript)

	taprootScriptTreeHash := taprootScriptTree.RootNode.TapHash()

	outputKey := txscript.ComputeTaprootOutputKey(internalKey, taprootScriptTreeHash[:])

	return txscript.PayToTaprootScript(outputKey)
}
