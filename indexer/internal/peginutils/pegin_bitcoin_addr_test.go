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
	input := struct {
		internalKey  string
		recoveryKey  string
		receiverAddr string
		csvLock      uint32
	}{
		internalKey:  "7eb791015af95721e8661f0de77250caa6b9a81044582096a639ab0bcbc268ef",
		recoveryKey:  "6f27fc5643a15d896df18c245436451a3a4584748e28c50a25acbaab7833d705",
		receiverAddr: "0:2eba6c129806cbb5d885c7a9dbf55bc5be72da5cd682ccb47a3b8d17dc57501d",
		csvLock:      29,
	}

	expectedAddr := "tb1prg8469u2nu4uqurudkagmddea7rfrfarsu2sdegd0n4j3kd9l4xss65etp"

	internalKey, _ := schnorr.ParsePubKey(utils.MustHexToBytes(input.internalKey, 32))
	recoveryKey, _ := schnorr.ParsePubKey(utils.MustHexToBytes(input.recoveryKey, 32))
	receiverAddr := address.MustParseRawAddr(input.receiverAddr)

	addr, err := CalcPeginBitcoinAddr(internalKey, recoveryKey, receiverAddr, input.csvLock)
	assert.NoError(t, err, "CalcPeginBitcoinAddr returned an error")
	assert.Equal(t, addr.String(), expectedAddr)
}

func TestBuildOpReturnScript(t *testing.T) {
	input := struct {
		addr string
	}{
		addr: "-1:d88d8dd15cabe68f1ced15dc80bec1098cac6933f9be174f82353c5e3d942059",
	}

	expectedScriptStr := "6a21d88d8dd15cabe68f1ced15dc80bec1098cac6933f9be174f82353c5e3d942059ff"
	addr := address.MustParseRawAddr(input.addr)

	script, err := buildOpReturnScript(addr)
	assert.NoError(t, err, "buildOpReturnScript returned an error")
	assert.Equal(t, expectedScriptStr, hex.EncodeToString(script))
}

func TestBuildCSVScript(t *testing.T) {
	input := struct {
		recoveryKey string
		csvLock     uint32
	}{
		recoveryKey: "5d81f9bb8895f1a76b969d5b6fa12c7f4f74b04e7588dc12f4cb775ee08d0f1e",
		csvLock:     127,
	}

	expectedScriptStr := "037f0040b275205d81f9bb8895f1a76b969d5b6fa12c7f4f74b04e7588dc12f4cb775ee08d0f1eac"
	recoveryKey, _ := schnorr.ParsePubKey(utils.MustHexToBytes(input.recoveryKey, 32))

	script, err := buildCSVScript(recoveryKey, input.csvLock)
	assert.NoError(t, err, "buildCSVScript returned an error")
	assert.Equal(t, expectedScriptStr, hex.EncodeToString(script))
}

func TestBuildTaprootScriptTree(t *testing.T) {
	input := struct {
		csvScript      string
		opReturnScript string
	}{
		csvScript:      "037f0040b275205d81f9bb8895f1a76b969d5b6fa12c7f4f74b04e7588dc12f4cb775ee08d0f1eac",
		opReturnScript: "6a21d88d8dd15cabe68f1ced15dc80bec1098cac6933f9be174f82353c5e3d942059ff",
	}

	expectedTreeHash := "73ce82f2f5e4b7033c205fbb642884ff33fd9b75c4a4cc7c5ef04b4f31f6df0a"
	csvScript, _ := hex.DecodeString(input.csvScript)
	opReturnScript, _ := hex.DecodeString(input.opReturnScript)

	tree := buildTaprootScriptTree(csvScript, opReturnScript)
	treeHash := tree.RootNode.TapHash()

	assert.Equal(t, expectedTreeHash, hex.EncodeToString(treeHash[:]))
}

func TestCalcPeginBitcoinAddrWithLeadingZeros(t *testing.T) {
	input := struct {
		internalKey  string
		receiverAddr string
		recoveryKey  string
		csvLock      uint32
	}{
		internalKey:  "593373d3531f0b7d07b687a0c549e2b7e5fe83128b42e1525a654cb8869c05f8",
		receiverAddr: "0:e412044eb05d492e29eb573dcf428e550e441d7f2dcac5a208bad25919e3d814",
		recoveryKey:  "00b17e025f5207507d77ab5ba1bb180fa18a5aa547b37b4621ee8c642d63aefd",
		csvLock:      338,
	}

	expectedAddr := "tb1p6ynhx4afr5vzk0ruwzjkq5g7guujy4z49j4zcz5z8jm67eg76hhq9d5kjx"

	internalKey, _ := schnorr.ParsePubKey(utils.MustHexToBytes(input.internalKey, 32))
	recoveryKey, _ := schnorr.ParsePubKey(utils.MustHexToBytes(input.recoveryKey, 32))
	receiverAddr := address.MustParseRawAddr(input.receiverAddr)

	addr, err := CalcPeginBitcoinAddr(internalKey, recoveryKey, receiverAddr, input.csvLock)
	assert.NoError(t, err, "CalcPeginBitcoinAddr returned an error")
	assert.Equal(t, addr.String(), expectedAddr)
}
