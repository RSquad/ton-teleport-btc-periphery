package peginutils

import (
	"encoding/hex"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/xssnick/tonutils-go/address"
)

func TestCalcPeginBitcoinAddr(t *testing.T) {
	expectedAddr := "tb1prg8469u2nu4uqurudkagmddea7rfrfarsu2sdegd0n4j3kd9l4xss65etp"

	internalKey, _ := schnorr.ParsePubKey(utils.MustHexToBytes("7eb791015af95721e8661f0de77250caa6b9a81044582096a639ab0bcbc268ef", 32))
	recoveryKey, _ := schnorr.ParsePubKey(utils.MustHexToBytes("6f27fc5643a15d896df18c245436451a3a4584748e28c50a25acbaab7833d705", 32))
	receiverAddr := address.MustParseRawAddr("0:2eba6c129806cbb5d885c7a9dbf55bc5be72da5cd682ccb47a3b8d17dc57501d")
	csvLock := uint32(29)

	addr, err := CalcPeginBitcoinAddr(internalKey, recoveryKey, receiverAddr, csvLock)
	assert.NoError(t, err, "CalcPeginBitcoinAddr returned an error")
	assert.Equal(t, addr.String(), expectedAddr)
}

func TestBuildOpReturnScript(t *testing.T) {
	expectedScriptStr := "6a21d88d8dd15cabe68f1ced15dc80bec1098cac6933f9be174f82353c5e3d942059ff"
	addr := address.MustParseRawAddr("-1:d88d8dd15cabe68f1ced15dc80bec1098cac6933f9be174f82353c5e3d942059")

	script, err := buildOpReturnScript(addr)
	assert.NoError(t, err, "buildOpReturnScript returned an error")
	assert.Equal(t, expectedScriptStr, hex.EncodeToString(script))
}

func TestBuildCSVScript(t *testing.T) {
	expectedScriptStr := "037f0040b275205d81f9bb8895f1a76b969d5b6fa12c7f4f74b04e7588dc12f4cb775ee08d0f1eac"
	recoveryKey, _ := schnorr.ParsePubKey(utils.MustHexToBytes("5d81f9bb8895f1a76b969d5b6fa12c7f4f74b04e7588dc12f4cb775ee08d0f1e", 32))
	csvLock := uint32(127)

	script, err := buildCSVScript(recoveryKey, csvLock)
	assert.NoError(t, err, "buildCSVScript returned an error")
	assert.Equal(t, expectedScriptStr, hex.EncodeToString(script))
}

func TestBuildTaprootScriptTree(t *testing.T) {
	expectedTreeHash := "73ce82f2f5e4b7033c205fbb642884ff33fd9b75c4a4cc7c5ef04b4f31f6df0a"
	csvScript, _ := hex.DecodeString("037f0040b275205d81f9bb8895f1a76b969d5b6fa12c7f4f74b04e7588dc12f4cb775ee08d0f1eac")
	opReturnScript, _ := hex.DecodeString("6a21d88d8dd15cabe68f1ced15dc80bec1098cac6933f9be174f82353c5e3d942059ff")

	tree := buildTaprootScriptTree(csvScript, opReturnScript)
	treeHash := tree.RootNode.TapHash()

	assert.Equal(t, expectedTreeHash, hex.EncodeToString(treeHash[:]))
}
