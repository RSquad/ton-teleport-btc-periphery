package bitcoin

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/pegoutcontract"
	tonclient "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/xssnick/tonutils-go/address"
)

func getClient() (*tonclient.TonClient, error) {
	return tonclient.New("https://ton-blockchain.github.io/testnet-global.config.json")
}

func TestBuildTaprootSigningHashes(t *testing.T) {
	type pegoutTestCase struct {
		name      string
		address   string
		expHashes []string
		expErr    bool
	}

	pegoutAddresses := []string{
		"0:51b405fc9bef127b8ab9448f2a640324b19679003e1126f898f8545d7d9f5029", // 1 input
		"0:2e930b5fe2d73785c749f146d8c1af9355a421a40bd4483a4ffe05cc9b1e4158", // 1 input
		"0:c4c5bfd2d3c9dbc37f2427322e81a599ab7267258ded5d9eca39e885bc2705a3", // 2 inputs
		"0:60c38094b17b52ada057249109c1d5ee3d2c73e16df3f0cfd9663dc741c77d93", // 52 inputs
	}

	testCases := []pegoutTestCase{
		{
			name:    "Pegout " + pegoutAddresses[0] + " (1 input)",
			address: pegoutAddresses[0],
			expHashes: []string{
				"4009b590d2f8ebef470cc994c947c58129ebb87bdc3d02c3061ecdcd7e62313e",
			},
			expErr: false,
		},
		{
			name:    "Pegout " + pegoutAddresses[1] + " (1 input)",
			address: pegoutAddresses[1],
			expHashes: []string{
				"07b797a5054be09ac3fa71461fdfb9c877b37bbe7a681e8fb932e5540cc0c64b",
			},
			expErr: false,
		},
		{
			name:    "Pegout " + pegoutAddresses[2] + " (2 inputs)",
			address: pegoutAddresses[2],
			expHashes: []string{
				"494183e26cb972d2bd453b174a8f1a309b9d3e87425d5fd18c068ac9f41527df",
				"7fd702d4c897345c5d4613cfb49fe7648b9cd08dd36e6a17cd75bd9c14f70f33",
			},
			expErr: false,
		},
		{
			name:    "Pegout " + pegoutAddresses[3] + " (52 inputs)",
			address: pegoutAddresses[3],
			expHashes: []string{
				"a41d80af5305bdc184de0a2938e4f080def82c7662456ba74ab67a7de25c4fab",
				"fec2e3b1e602fc3434c8098bc8b51824cc81b3f55fb95579cb4108f74bb36109",
				"ed38be566f2ecd74575b7bee423dff1d0cce2bf74290dd61465d83f636f36ad2",
				"9e525482a73173bbedbf8e1dbd27f83c9146fcac44af0664bd1549d95cb223bf",
				"1ca8f1d51697b035a14f2c60fac4e668b489a2545171050636abbe6ce6aac75b",
				"a929ee3e8eeff81b18f10be6d8009ded0256e235416883817673c58abb32b93c",
				"713588cc3cd5079761c4bb584d8c0698468179beb9d0458a33477fd1717ce0fc",
				"987954f545e468f1e0062498d0a9d7fbdfe1f4197f686a0e215069d267b129fe",
				"b615660a918df750654c0f4c26cb5360b7b6bbdd54d2f4ea69f9b092993c1fc1",
				"21b32fae120d8e9c00749f4133549e4016c85f5b8dfbbe2cc99f75f43237489d",
				"a3c3a84a7f5300ccb9dd0a3f0152ce6af91a0841403e35006444fc60beb3de53",
				"e8221a2e64d8dcb8d669e4141e8747afa28663be641cdb046ce9036fc44598f5",
				"4a58415d653738e1db26424d99f2516ee4acde073d49438d2929f5c45c2fc039",
				"7dd35d0df23c3b240395503c5a8a49221d11efc9c35d1b91838c97da544b62a6",
				"5deeaa26df401482b74a80eb74b076ffe663cad104af595db25d7c656a403be7",
				"144a7d720fa95671c9f25cad530f8237804349a4c62abe60c6d8766f741c328e",
				"fe02482eafc2c61f591c70cf0c845e53dc84b2850db9d0b1ec0b3b473022a9a9",
				"142e19796f8a183cccddbd55168ebd4f85291b202b5cd56558bba84dce73c1fd",
				"0c6d51d7d9988fab23d66eb1ec146d29ba4700e7a3522f17974904faf2ede559",
				"ea63468f3c20189f6e4c431997f4debe3c569a6cfe056e8cf48ad2004fa8c8a7",
				"d7e64ea544fe0d5a6f3f6f5b2cc759a5fc7d8fec623fb03923725513c6bddcaa",
				"ae1b8a8c75f8c8d93f6ae3cda8214b1988619502916c4f3be4f8a4a62248db8f",
				"6004353647812aac9ebd1f1e723ace28fac63fbd4ba3c51eae61810cfb816c22",
				"a2f6b816edad79a29b0fb247801d6297714df9efbe4b4d42a887f6de58be3af2",
				"89c5e3aa9f26a6f3f2b1de2b24a1618f953ddc64668e159f7bf19dcd98ee9405",
				"88e3814fdfc8226ef55ef30948beccba29f667effdc96f13195b79b0173a6cb7",
				"ece957e94aa3098f15c0f4001ec19ddda849e5adfa685e7b6de47494eb9c7979",
				"f04d4481c54f326b56f8001dac62387504f21daef125862ac8a350042c5f8351",
				"55a3c69a4d65948e7bcea274d36ed51dbf4bfb090e50ffc2d61df7ef299f6877",
				"ba12089600f13450b4d508ec38f136cad11ab0e9591f5d265e4d6222c7292b2b",
				"638cc7b047242b24e1f40167d435d6322d4d0d19038d3feb277b78c6c6333913",
				"fcbb3de1496fa84e8743a723b6c15bbba5f94ff31b3fbf698387f75b9ec64bb2",
				"0b0e148169ec00d01f7fe288cbb0369eeb3f32f5dac15d41d13a6164ee1dcd1f",
				"1aa421a1249a0a5071d26269e7b31c5708c0c89e4be81e159133b8290f98fd23",
				"f145aebcb5b340ff534994f1fab2714a258c0306460b0754cd644f96eb265cd8",
				"94c5f6b626892684aa3d8ef56d337d23f40d769cbd4a00d9fe4a75975f11543d",
				"0263ddf3ffc244066630b0ccff81d87156d9734a6b18b022c17a663f5bef882d",
				"5c36b75a2753a9c61178a21f84b108045de700bbc21e018d18b9f375919cf5d8",
				"e991995629fe21d300605147e5c3ba40698ff4b4445c090ff9e148289eba75dd",
				"084e699d60b106f682949b179df780b9698b084a592e3c0e02295949e5117419",
				"801d4e8d88a505b6f0bcbe8f123a33f5d2f5f23c8e3d0c7cbc2d4d74a73fb59b",
				"3fab86fc009a21306633a1e10104312a685329a00d35d3ff9477505b65387d91",
				"d6baddf763f5d6e2069714aa80aebbccf6c410e9bf117af40cac53e1d30c3f01",
				"461944ec13aaeda461f5d74320a9ac72fc56bd5b99155f233979acc5be9ef719",
				"4662be0b1e39c9b4acba8910cc72c436200f406579bcdb34e522098879d1c947",
				"d69820fa5ac93d0187c18e845f336aaeb5cf675ff72f82cdcab46240b8ec539f",
				"d79100da48641a9d979e48bbd4d15622fca667a1450d27e2a4045e6166ac3d49",
				"d37e90c019867d513ba8bd1075859a21679d61693d19ce9a795564dd1ce91564",
				"2acdb86a24d85d688eb2bd7139274042501a6bad1333691948fae3df805eef12",
				"0605910792ec4fd467b3421e440a60b632c32bf75e5ba63e16a19a539d1a434d",
				"44c8806e03a5ab1a0d231845b91d675daf5f2014b0fc363c7aed1e47f37cd050",
				"40e981e71fefc000acf1cf263770a32bad86ef426c747d069a22f15f852a975e",
			},
			expErr: false,
		},
	}

	client, err := getClient()
	if err != nil {
		t.Fatalf("Failed to get tonclient for tests: %v", err)
	}
	ctx := context.Background()

	block, err := client.API.CurrentMasterchainInfo(ctx)
	if err != nil {
		t.Errorf("call CurrentMasterchainInfo failed: %e", err)
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			addr := address.MustParseRawAddr(tc.address)

			pegoutContract := pegoutcontract.New(addr, client, ctx)
			txParts, err := pegoutContract.GetTxParts(block)
			if err != nil {
				t.Fatalf("GetTxParts failed for address %s: %v", tc.address, err)
			}

			hashes, err := BuildTaprootSigningHashes(*txParts.Inputs, txParts.Outputs)

			if tc.expErr {
				if err == nil {
					t.Fatalf("expected an error from BuildTaprootSigningHashes, but got nil for address %s", tc.address)
				}
				return
			}
			if err != nil {
				t.Fatalf("BuildTaprootSigningHashes failed unexpectedly for address %s: %v", tc.address, err)
			}

			if len(hashes) != len(tc.expHashes) {
				t.Fatalf("expected %d hashes, got %d for address %s", len(tc.expHashes), len(hashes), tc.address)
			}

			for i, h := range hashes {
				hashStr := hex.EncodeToString(h)
				if hashStr != tc.expHashes[i] {
					t.Errorf("hash %d mismatch for address %s: expected %s, got %s", i, tc.address, tc.expHashes[i], hashStr)
				}
			}
		})
	}
}
