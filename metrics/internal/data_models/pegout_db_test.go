package data_models

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xssnick/tonutils-go/address"
)

func TestPegoutStatus_MarshalText(t *testing.T) {
	tests := []struct {
		name        string
		status      PegoutStatus
		expected    string
		expectError bool
	}{
		{"confirmed", PEGOUT_CONFIRMED, "CONFIRMED", false},
		{"signed", PEGOUT_SIGNED, "SIGNED", false},
		{"signing", PEGOUT_SIGNING, "SIGNING", false},
		{"invalid", PegoutStatus(999), "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, err := tt.status.MarshalText()
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(text))
		})
	}
}

func TestPegoutStatus_UnmarshalText(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    PegoutStatus
		expectError bool
	}{
		{"confirmed", "CONFIRMED", PEGOUT_CONFIRMED, false},
		{"signed", "SIGNED", PEGOUT_SIGNED, false},
		{"signing", "SIGNING", PEGOUT_SIGNING, false},
		{"lowercase", "confirmed", 0, true},
		{"invalid", "INVALID", 0, true},
		{"empty", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var status PegoutStatus
			err := status.UnmarshalText([]byte(tt.input))
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, status)
		})
	}

	t.Run("nil receiver", func(t *testing.T) {
		var status *PegoutStatus
		err := status.UnmarshalText([]byte("CONFIRMED"))
		assert.Error(t, err)
	})
}

func TestDeserializePegoutDB(t *testing.T) {
	validTonAddr := "0:dc3e60e9b67fb4fa05801ae6e530477b4238053ed8d76d085c2a4c016139f354"
	validTonAddr2 := "0:fc300993a58493fcd758007dd38ca9681f6e5a90026d1d04464481cb7136d51e"
	txRaw := "020000000001029ba02bc0d516f547524c505c19fc08b1d15dc40876517547c81e4d2f608a9a3d0100000000fdfffffffc1ce5c729036c9f993249d3b7ca9aff4699bb530b35816cb65002cecbb2ce7d0000000000fdffffff01fb76b41f63000000225120b7b24f6f535ecd06d30808de0a76a11f5bb7d5d9ccf6abf80bab8489c01cc04f01401f6ab1f92872377427fdeaaf5c45cbbbd1c4131436b537ff08a63f373ec45723e69e8c7abbb0e6e468bc95d94580fb6efb55a1c3ea72f0642d4828a7eadb620c014013832bfecd3f7f9118891a9af9eced445e805f4c6d93a3891c06c61627db9929250f25da5cec59ad361613882b2948c7495a3656701e21ba746009fa9eed19cd00000000"
	txID := "c818e8cc02acabc5863b4ca69acb6d00d04ab30df79e151014a35d1216369586"
	blockHash := ""

	validJSON := fmt.Sprintf(`{
		"id": 1,
		"addr": "%s",
		"status": "CONFIRMED",
		"bitcoin_tx_raw": "%s",
		"bitcoin_tx_id": "%s",
		"bitcoin_block_hash": "%s"
	}`, validTonAddr, txRaw, txID, blockHash)

	validJSONNumericStatus := fmt.Sprintf(`{
		"id": 2,
		"addr": "%s",
		"status": 1,
		"bitcoin_tx_raw": "%s",
		"bitcoin_tx_id": "%s",
		"bitcoin_block_hash": "%s"
	}`, validTonAddr2, txRaw, txID, blockHash)

	tests := []struct {
		name        string
		jsonData    string
		address     string
		expectError bool
		checkFunc   func(*testing.T, *Pegout)
	}{
		{
			name:     "valid json with string status",
			jsonData: validJSON,
			address:  "test-address",
			checkFunc: func(t *testing.T, p *Pegout) {
				assert.Equal(t, uint64(1), p.Id)
				assert.Equal(t, PEGOUT_CONFIRMED, p.Status)
				assert.NotNil(t, p.Addr)
				tonAddr := (*address.Address)(p.Addr)
				assert.Equal(t, validTonAddr, tonAddr.StringRaw())
				assert.Equal(t, txRaw, hex.EncodeToString(p.BitcoinTxRaw))
				assert.Equal(t, txID, hex.EncodeToString(p.BitcoinTxId))
				assert.Equal(t, blockHash, hex.EncodeToString(p.BitcoinBlockHash))
			},
		},
		{
			name:     "valid json with numeric status",
			jsonData: validJSONNumericStatus,
			address:  "test-address",
			checkFunc: func(t *testing.T, p *Pegout) {
				assert.Equal(t, uint64(2), p.Id)
				assert.Equal(t, PEGOUT_SIGNED, p.Status)
				assert.NotNil(t, p.Addr)
				tonAddr := (*address.Address)(p.Addr)
				assert.Equal(t, validTonAddr2, tonAddr.StringRaw())
			},
		},
		{
			name:        "invalid json",
			jsonData:    `{invalid json}`,
			address:     "test-address",
			expectError: true,
		},
		{
			name: "json with empty hex fields",
			jsonData: fmt.Sprintf(`{
				"id": 3,
				"addr": "%s",
				"status": "SIGNING",
				"bitcoin_tx_raw": "",
				"bitcoin_tx_id": "",
				"bitcoin_block_hash": ""
			}`, validTonAddr),
			address: "test-address",
			checkFunc: func(t *testing.T, p *Pegout) {
				assert.Equal(t, uint64(3), p.Id)
				assert.Equal(t, PEGOUT_SIGNING, p.Status)
				assert.Nil(t, p.BitcoinTxRaw)
				assert.Nil(t, p.BitcoinTxId)
				assert.Nil(t, p.BitcoinBlockHash)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pegout, err := DeserializePegoutDB([]byte(tt.jsonData), tt.address)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.address)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, pegout)
			if tt.checkFunc != nil {
				tt.checkFunc(t, pegout)
			}
		})
	}
}

func TestDeserializePegoutsDB(t *testing.T) {
	validTonAddr := "0:dc3e60e9b67fb4fa05801ae6e530477b4238053ed8d76d085c2a4c016139f354"
	txRaw := "020000000001029ba02bc0d516f547524c505c19fc08b1d15dc40876517547c81e4d2f608a9a3d0100000000fdfffffffc1ce5c729036c9f993249d3b7ca9aff4699bb530b35816cb65002cecbb2ce7d0000000000fdffffff01fb76b41f63000000225120b7b24f6f535ecd06d30808de0a76a11f5bb7d5d9ccf6abf80bab8489c01cc04f01401f6ab1f92872377427fdeaaf5c45cbbbd1c4131436b537ff08a63f373ec45723e69e8c7abbb0e6e468bc95d94580fb6efb55a1c3ea72f0642d4828a7eadb620c014013832bfecd3f7f9118891a9af9eced445e805f4c6d93a3891c06c61627db9929250f25da5cec59ad361613882b2948c7495a3656701e21ba746009fa9eed19cd00000000"
	txID := "c818e8cc02acabc5863b4ca69acb6d00d04ab30df79e151014a35d1216369586"
	blockHash := ""

	validJSONArray := fmt.Sprintf(`[
		{
			"id": 1,
			"addr": "%s",
			"status": "CONFIRMED",
			"bitcoin_tx_raw": "%s",
			"bitcoin_tx_id": "%s",
			"bitcoin_block_hash": "%s"
		},
		{
			"id": 2,
			"addr": "%s",
			"status": "SIGNED",
			"bitcoin_tx_raw": "%s",
			"bitcoin_tx_id": "%s",
			"bitcoin_block_hash": "%s"
		}
	]`, validTonAddr, txRaw, txID, blockHash, validTonAddr, txRaw, txID, blockHash)

	tests := []struct {
		name        string
		jsonData    string
		address     string
		expectError bool
		expectedLen int
	}{
		{
			name:        "valid json array",
			jsonData:    validJSONArray,
			address:     "test-address",
			expectedLen: 2,
		},
		{
			name:        "empty array",
			jsonData:    `[]`,
			address:     "test-address",
			expectedLen: 0,
		},
		{
			name:        "invalid json",
			jsonData:    `[invalid json]`,
			address:     "test-address",
			expectError: true,
		},
		{
			name: "array with invalid element",
			jsonData: fmt.Sprintf(`[
				{
					"id": 1,
					"addr": "%s",
					"status": "CONFIRMED",
					"bitcoin_tx_raw": "%s",
					"bitcoin_tx_id": "%s",
					"bitcoin_block_hash": "%s"
				},
				{
					"id": 2,
					"addr": "%s",
					"status": "INVALID_STATUS",
					"bitcoin_tx_raw": "%s",
					"bitcoin_tx_id": "%s",
					"bitcoin_block_hash": "%s"
				}
			]`, validTonAddr, txRaw, txID, blockHash, validTonAddr, txRaw, txID, blockHash),
			address:     "test-address",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pegouts, err := DeserializePegoutsDB([]byte(tt.jsonData))
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, pegouts)
			assert.Len(t, pegouts, tt.expectedLen)

			if tt.expectedLen > 0 {
				// Проверяем первый элемент
				assert.Equal(t, uint64(1), pegouts[0].Id)
				assert.Equal(t, PEGOUT_CONFIRMED, pegouts[0].Status)
				tonAddr := (*address.Address)(pegouts[0].Addr)
				assert.Equal(t, validTonAddr, tonAddr.StringRaw())
			}
		})
	}
}

func TestPegoutJsonToPegout(t *testing.T) {
	validTonAddr := "0:dc3e60e9b67fb4fa05801ae6e530477b4238053ed8d76d085c2a4c016139f354"
	addr, _ := address.ParseRawAddr(validTonAddr)
	pegoutTonAddr := (*PegoutTonAddr)(addr)

	tests := []struct {
		name        string
		pegoutJson  *PegoutJSON
		expectError bool
		checkFunc   func(*testing.T, *Pegout)
	}{
		{
			name: "complete data",
			pegoutJson: &PegoutJSON{
				Id:                  1,
				Addr:                pegoutTonAddr,
				Status:              PEGOUT_CONFIRMED,
				BitcoinTxRawHex:     "010203",
				BitcoinTxIdHex:      "040506",
				BitcoinBlockHashHex: "070809",
			},
			checkFunc: func(t *testing.T, p *Pegout) {
				assert.Equal(t, uint64(1), p.Id)
				assert.Equal(t, pegoutTonAddr, p.Addr)
				assert.Equal(t, PEGOUT_CONFIRMED, p.Status)
				assert.Equal(t, []byte{1, 2, 3}, p.BitcoinTxRaw)
				assert.Equal(t, []byte{4, 5, 6}, p.BitcoinTxId)
				assert.Equal(t, []byte{7, 8, 9}, p.BitcoinBlockHash)
			},
		},
		{
			name: "empty hex fields",
			pegoutJson: &PegoutJSON{
				Id:                  2,
				Addr:                pegoutTonAddr,
				Status:              PEGOUT_SIGNING,
				BitcoinTxRawHex:     "",
				BitcoinTxIdHex:      "",
				BitcoinBlockHashHex: "",
			},
			checkFunc: func(t *testing.T, p *Pegout) {
				assert.Equal(t, uint64(2), p.Id)
				assert.Equal(t, PEGOUT_SIGNING, p.Status)
				assert.Nil(t, p.BitcoinTxRaw)
				assert.Nil(t, p.BitcoinTxId)
				assert.Nil(t, p.BitcoinBlockHash)
			},
		},
		{
			name: "hex with 0x prefix",
			pegoutJson: &PegoutJSON{
				Id:                  3,
				Addr:                pegoutTonAddr,
				Status:              PEGOUT_SIGNED,
				BitcoinTxRawHex:     "0x0A0B0C",
				BitcoinTxIdHex:      "0x0D0E0F",
				BitcoinBlockHashHex: "0x101112",
			},
			checkFunc: func(t *testing.T, p *Pegout) {
				assert.Equal(t, []byte{10, 11, 12}, p.BitcoinTxRaw)
				assert.Equal(t, []byte{13, 14, 15}, p.BitcoinTxId)
				assert.Equal(t, []byte{16, 17, 18}, p.BitcoinBlockHash)
			},
		},
		{
			name: "hex with uppercase",
			pegoutJson: &PegoutJSON{
				Id:                  4,
				Addr:                pegoutTonAddr,
				Status:              PEGOUT_CONFIRMED,
				BitcoinTxRawHex:     "0X0A0B0C",
				BitcoinTxIdHex:      "ABCDEF",
				BitcoinBlockHashHex: "123456",
			},
			checkFunc: func(t *testing.T, p *Pegout) {
				assert.Equal(t, []byte{10, 11, 12}, p.BitcoinTxRaw)
				assert.Equal(t, []byte{0xAB, 0xCD, 0xEF}, p.BitcoinTxId)
				assert.Equal(t, []byte{0x12, 0x34, 0x56}, p.BitcoinBlockHash)
			},
		},
		{
			name: "invalid hex",
			pegoutJson: &PegoutJSON{
				Id:                  5,
				Addr:                pegoutTonAddr,
				Status:              PEGOUT_CONFIRMED,
				BitcoinTxRawHex:     "invalid-hex",
				BitcoinTxIdHex:      "040506",
				BitcoinBlockHashHex: "070809",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pegout, err := PegoutJsonToPegout(tt.pegoutJson)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, pegout)
			if tt.checkFunc != nil {
				tt.checkFunc(t, pegout)
			}
		})
	}
}
