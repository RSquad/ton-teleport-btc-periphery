package bitcoin

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/pegoutcontract"
	"github.com/xssnick/tonutils-go/address"
)

func stob(s string) []byte {
	bytes, _ := hex.DecodeString(s)
	return bytes
}

func TestBuildTaprootSigningHashes(t *testing.T) {
	// pegout 0:51b405fc9bef127b8ab9448f2a640324b19679003e1126f898f8545d7d9f5029
	inputs := pegoutcontract.TxPartsInputs{
		"a693bb75d008170d66a512e4ff96f1cdd40ca30d1e029389a9fad2d702ce641b": {
			Amount:            big.NewInt(25520022421),
			Index:             1,
			Receiver:          address.MustParseAddr("EQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAM9c"),
			BitcoinMerkleRoot: make([]byte, 0, 32),
			BitcoinScript:     stob("51207d632a944d845e3904c4f342819d541bc42492b5ade36502b56a264526006e3f"),
		},
	}
	outputs := []*pegoutcontract.TxPartsOutput{
		{
			Amount:        30000,
			BitcoinScript: stob("001413c3e3a88837b02c2453cab5e786c7d8679f398f"),
		},
		{
			Amount:        25519947421,
			BitcoinScript: stob("51207d632a944d845e3904c4f342819d541bc42492b5ade36502b56a264526006e3f"),
		},
	}

	expHashes := []string{
		"4009b590d2f8ebef470cc994c947c58129ebb87bdc3d02c3061ecdcd7e62313e",
	}

	hashes, err := BuildTaprootSigningHashes(inputs, outputs)
	if err != nil {
		t.Fatalf("failed to build taproot signing hashes: %v", err)
	}

	if len(hashes) != len(expHashes) {
		t.Fatalf("expected %d hashes, got %d", len(expHashes), len(hashes))
	}

	for i, h := range hashes {
		hashStr := hex.EncodeToString(h)
		if hashStr != expHashes[i] {
			t.Errorf("hash %d mismatch: expected %s, got %s", i, expHashes[i], hashStr)
		}
	}
}
