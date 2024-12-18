package peginutils

import (
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/xssnick/tonutils-go/address"
)

func CalcPeginBitcoinAddr(
	internalKey *btcec.PublicKey,
	recoveryKey *btcec.PublicKey,
	receiverAddr *address.Address,
	teleportContractStorage *teleportcontract.Storage,
) (*btcutil.AddressTaproot, error) {
	csvLock := teleportContractStorage.CsvLock

	csvScript, _ := txscript.NewScriptBuilder().
		AddInt64((1 << 22) | csvLock).
		AddOp(txscript.OP_CHECKSEQUENCEVERIFY).
		AddOp(txscript.OP_DROP).
		AddData(recoveryKey.X().Bytes()).
		AddOp(txscript.OP_CHECKSIG).
		Script()

	opReturnScript, _ := txscript.NewScriptBuilder().
		AddOp(txscript.OP_RETURN).
		AddData(bitcoin.TonAddrToBytesForTapLeaf(receiverAddr)).
		Script()

	taprootScriptTree := txscript.AssembleTaprootScriptTree(
		txscript.NewBaseTapLeaf(csvScript),
		txscript.NewBaseTapLeaf(opReturnScript),
	)

	rootHash := taprootScriptTree.RootNode.TapHash()

	outputKey := txscript.ComputeTaprootOutputKey(internalKey, rootHash[:])

	return btcutil.NewAddressTaproot(
		schnorr.SerializePubKey(outputKey),
		&chaincfg.SigNetParams,
	)
}
