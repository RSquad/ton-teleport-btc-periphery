package data_models

import (
	"testing"

	"github.com/xssnick/tonutils-go/address"
)

func TestDeserializePegoutDB_Success(t *testing.T) {
	jsonData := `{"id":17179871213,"addr":"0:5804cd4a2124ee151bef2143782d6e7ca245242a61994e8049826e5ab0fe5f15","status":"SIGNED","bitcoin_tx_raw":"02000000000101c2deebcdd3fa884b9884d3fbdb621e81e1b8f2c794083d980821d51ceafe12100100000000ffffffff02a90b00000000000016001440028208c5734dc80d0f430943b51322590da6534df9161608000000225120c36eb5c0b6e9f2531f86fd2a4d920d47a44c19f6405e30391fbb57ae88f7139401403276b21bb326a9784389d7102627f31d9476b0008c474f115a73a66dcebce7011f134bf893cadfb8214df9fff337e63203d5b432e2c801070a009f115f6f38d500000000","bitcoin_tx_id":"0340d912c94896d82ecb6df1aa9ba76ff9d19a69de1efe27c89b74fa61e43e34","bitcoin_block_hash":""}`

	pegout, err := DeserializePegoutDB([]byte(jsonData))
	if err != nil {
		t.Fatalf("DeserializePegouts error: %v", err)
		return
	}

	addr, err := address.ParseRawAddr("0:5804cd4a2124ee151bef2143782d6e7ca245242a61994e8049826e5ab0fe5f15")
	if err != nil {
		t.Fatalf("DeserializePegouts error: %v", err)
		return
	}

	if pegout.Id != 17179871213 {
		t.Fatal("Wrong address")
		return
	}

	if (*address.Address)(pegout.Addr).StringRaw() != addr.StringRaw() {
		t.Fatal("Wrong address")
		return
	}

}
