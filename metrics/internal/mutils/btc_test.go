package mutils

import (
	"errors"
	"net/http"
	"testing"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBitcoinClient is a mock implementation of bitcoin.Client
// MockBitcoinClient implements bitcoin.Client interface for testing
// MockBitcoinClient is a mock implementation of bitcoin.Client
type MockBitcoinClient struct {
	mock.Mock
}

// Ensure MockBitcoinClient implements bitcoin.Client interface
var _ bitcoin.ClientInterface = (*MockBitcoinClient)(nil)

// GetBlockCount returns the current block count
func (m *MockBitcoinClient) GetBlockCount() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

// GetBlockHash returns the hash of a block at the given height
func (m *MockBitcoinClient) GetBlockHash(blockHeight int64) (*chainhash.Hash, error) {
	args := m.Called(blockHeight)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*chainhash.Hash), args.Error(1)
}

// GetBlockHeaderVerbose returns the block header for the given hash
func (m *MockBitcoinClient) GetBlockHeaderVerbose(hash *chainhash.Hash) (*btcjson.GetBlockHeaderVerboseResult, error) {
	args := m.Called(hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*btcjson.GetBlockHeaderVerboseResult), args.Error(1)
}

// GetBlockChainInfo returns blockchain information
func (m *MockBitcoinClient) GetBlockChainInfo() (*btcjson.GetBlockChainInfoResult, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*btcjson.GetBlockChainInfoResult), args.Error(1)
}

// GetRawTransactionVerbose returns verbose transaction information
func (m *MockBitcoinClient) GetRawTransactionVerbose(txHash *chainhash.Hash) (*btcjson.TxRawResult, error) {
	args := m.Called(txHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*btcjson.TxRawResult), args.Error(1)
}

// GetBlockHeightByHash returns the block height for the given hash
func (m *MockBitcoinClient) GetBlockHeightByHash(hash *chainhash.Hash) (int64, error) {
	args := m.Called(hash)
	return args.Get(0).(int64), args.Error(1)
}

// GetBlockHashByTxID returns the block hash for the given transaction ID
func (m *MockBitcoinClient) GetBlockHashByTxID(txID *chainhash.Hash) (*chainhash.Hash, error) {
	args := m.Called(txID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*chainhash.Hash), args.Error(1)
}

// GetTxProof returns the transaction proof for the given transaction and block
func (m *MockBitcoinClient) GetTxProof(txID *chainhash.Hash, blockHash *chainhash.Hash) (string, error) {
	args := m.Called(txID, blockHash)
	return args.String(0), args.Error(1)
}

// GetBlockHashesByStartHeight returns block hashes starting from the given height
func (m *MockBitcoinClient) GetBlockHashesByStartHeight(startHeight int64, count int64) ([]*chainhash.Hash, error) {
	args := m.Called(startHeight, count)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*chainhash.Hash), args.Error(1)
}

// GetTxChildrenCount returns the children count for a transaction
func (m *MockBitcoinClient) GetTxChildrenCount(parentHash *chainhash.Hash) (*bitcoin.TxChildrenCount, error) {
	args := m.Called(parentHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*bitcoin.TxChildrenCount), args.Error(1)
}

// EstimateFee estimates the transaction fee
func (m *MockBitcoinClient) EstimateFee(blockCount int, estimateMode *btcjson.EstimateSmartFeeMode) (float64, error) {
	args := m.Called(blockCount, estimateMode)
	return args.Get(0).(float64), args.Error(1)
}

// SendRawTransaction sends a raw transaction
func (m *MockBitcoinClient) SendRawTransaction(tx *wire.MsgTx, allowHighFees bool) (*chainhash.Hash, error) {
	args := m.Called(tx, allowHighFees)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*chainhash.Hash), args.Error(1)
}

// GetTxOut returns transaction output details
func (m *MockBitcoinClient) GetTxOut(txid *chainhash.Hash, vout uint32, includeMempool bool) (*btcjson.GetTxOutResult, error) {
	args := m.Called(txid, vout, includeMempool)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*btcjson.GetTxOutResult), args.Error(1)
}

func (m *MockBitcoinClient) GetRPCClient() *rpcclient.Client {
	args := m.Called()
	return args.Get(0).(*rpcclient.Client)
}

func (m *MockBitcoinClient) GetHTTPClient() *http.Client {
	args := m.Called()
	return args.Get(0).(*http.Client)
}

// ShutdownRPCClient shuts down the RPC client
func (m *MockBitcoinClient) ShutdownRPCClient() {
	m.Called()
}

// Now the test functions should work without interface assignment issues
func TestBtcGetBestBlockHeight(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockBitcoinClient)
		expectedHeight int
		expectedError  string
	}{
		{
			name: "Success - returns block height",
			setupMock: func(mockClient *MockBitcoinClient) {
				info := &btcjson.GetBlockChainInfoResult{
					Blocks: 850000,
				}
				mockClient.On("GetBlockChainInfo").Return(info, nil)
			},
			expectedHeight: 850000,
			expectedError:  "",
		},
		{
			name: "Error - GetBlockChainInfo fails",
			setupMock: func(mockClient *MockBitcoinClient) {
				mockClient.On("GetBlockChainInfo").Return(nil, errors.New("rpc error"))
			},
			expectedHeight: 0,
			expectedError:  "failed to get btc blockchain info: rpc error",
		},
		{
			name: "Success - zero blocks",
			setupMock: func(mockClient *MockBitcoinClient) {
				info := &btcjson.GetBlockChainInfoResult{
					Blocks: 0,
				}
				mockClient.On("GetBlockChainInfo").Return(info, nil)
			},
			expectedHeight: 0,
			expectedError:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockBitcoinClient)
			tt.setupMock(mockClient)

			height, err := BtcGetBestBlockHeight(mockClient)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedHeight, height)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestBtcGetCPFPChainSize(t *testing.T) {
	// Define all transaction hash variables
	const (
		tx1 = "a1075db55d416d3ca199f55b6084e2115b9345e16c5cf302fc80e9d5fbf5d48d"
		tx2 = "b20858b4d416d3ca199f55b6084e2115b9345e16c5cf302fc80e9d5fbf5d48d1"
		tx3 = "c30967a4d416d3ca199f55b6084e2115b9345e16c5cf302fc80e9d5fbf5d48d2"
		tx4 = "d40a58b4d416d3ca199f55b6084e2115b9345e16c5cf302fc80e9d5fbf5d48d3"
		tx5 = "e50b59b5d416d3ca199f55b6084e2115b9345e16c5cf302fc80e9d5fbf5d48d4"
		tx6 = "f60c5ab6d416d3ca199f55b6084e2115b9345e16c5cf302fc80e9d5fbf5d48d5"
	)

	// Helper function to create proper Bitcoin transaction hashes
	createHash := func(s string) *chainhash.Hash {
		hash, err := chainhash.NewHashFromStr(s)
		if err != nil {
			panic(err)
		}
		return hash
	}

	// Helper function to create TxRawResult
	createTxResult := func(txid string, confirmations uint64, vin []btcjson.Vin) *btcjson.TxRawResult {
		return &btcjson.TxRawResult{
			Txid:          txid,
			Confirmations: confirmations,
			Vin:           vin,
		}
	}

	// Helper function to create Vin
	createVin := func(txid string) btcjson.Vin {
		return btcjson.Vin{
			Txid: txid,
		}
	}

	tests := []struct {
		name          string
		txHash        *chainhash.Hash
		setupMock     func(*MockBitcoinClient)
		expectedSize  int
		expectedError string
	}{
		{
			name:   "Success - single unconfirmed transaction",
			txHash: createHash(tx1),
			setupMock: func(mockClient *MockBitcoinClient) {
				// Current transaction is unconfirmed
				mockClient.On("GetRawTransactionVerbose", createHash(tx1)).
					Return(createTxResult(tx1, 0, []btcjson.Vin{
						createVin(tx2),
					}), nil)

				// Parent transaction is confirmed
				mockClient.On("GetRawTransactionVerbose", createHash(tx2)).
					Return(createTxResult(tx2, 1, []btcjson.Vin{}), nil)
			},
			expectedSize:  1,
			expectedError: "",
		},
		{
			name:   "Success - confirmed transaction returns 0",
			txHash: createHash(tx3),
			setupMock: func(mockClient *MockBitcoinClient) {
				mockClient.On("GetRawTransactionVerbose", createHash(tx3)).
					Return(createTxResult(tx3, 5, []btcjson.Vin{}), nil)
			},
			expectedSize:  0,
			expectedError: "",
		},
		{
			name:   "Success - chain of 3 unconfirmed transactions",
			txHash: createHash(tx4),
			setupMock: func(mockClient *MockBitcoinClient) {
				// Transaction 4 (current)
				mockClient.On("GetRawTransactionVerbose", createHash(tx4)).
					Return(createTxResult(tx4, 0, []btcjson.Vin{
						createVin(tx5),
					}), nil)

				// Transaction 5 (parent)
				mockClient.On("GetRawTransactionVerbose", createHash(tx5)).
					Return(createTxResult(tx5, 0, []btcjson.Vin{
						createVin(tx6),
					}), nil)

				// Transaction 6 (grandparent)
				mockClient.On("GetRawTransactionVerbose", createHash(tx6)).
					Return(createTxResult(tx6, 0, []btcjson.Vin{
						createVin(tx1),
					}), nil)

				// Confirmed transaction (stops the chain)
				mockClient.On("GetRawTransactionVerbose", createHash(tx1)).
					Return(createTxResult(tx1, 10, []btcjson.Vin{}), nil)
			},
			expectedSize:  3,
			expectedError: "",
		},
		{
			name:   "Success - multiple parents but only one unconfirmed",
			txHash: createHash(tx2),
			setupMock: func(mockClient *MockBitcoinClient) {
				mockClient.On("GetRawTransactionVerbose", createHash(tx2)).
					Return(createTxResult(tx2, 0, []btcjson.Vin{
						createVin(tx3), // confirmed parent
						createVin(tx4), // unconfirmed parent
					}), nil)

				// First parent is confirmed
				mockClient.On("GetRawTransactionVerbose", createHash(tx3)).
					Return(createTxResult(tx3, 5, []btcjson.Vin{}), nil)

				// Second parent is unconfirmed
				mockClient.On("GetRawTransactionVerbose", createHash(tx4)).
					Return(createTxResult(tx4, 0, []btcjson.Vin{
						createVin(tx5),
					}), nil)

				// Grandparent is confirmed
				mockClient.On("GetRawTransactionVerbose", createHash(tx5)).
					Return(createTxResult(tx5, 1, []btcjson.Vin{}), nil)
			},
			expectedSize:  2,
			expectedError: "",
		},
		{
			name:   "Success - coinbase transaction in chain",
			txHash: createHash(tx6),
			setupMock: func(mockClient *MockBitcoinClient) {
				mockClient.On("GetRawTransactionVerbose", createHash(tx6)).
					Return(createTxResult(tx6, 0, []btcjson.Vin{
						{Txid: ""}, // Coinbase transaction (empty Txid)
						createVin(tx1),
					}), nil)

				mockClient.On("GetRawTransactionVerbose", createHash(tx1)).
					Return(createTxResult(tx1, 1, []btcjson.Vin{}), nil)
			},
			expectedSize:  1,
			expectedError: "",
		},
		{
			name:   "Error - GetRawTransactionVerbose fails for current tx",
			txHash: createHash(tx1),
			setupMock: func(mockClient *MockBitcoinClient) {
				mockClient.On("GetRawTransactionVerbose", createHash(tx1)).
					Return(nil, errors.New("rpc error"))
			},
			expectedSize:  0,
			expectedError: "failed to get transaction: rpc error",
		},
		{
			name:   "Error - GetRawTransactionVerbose not fails for parent tx",
			txHash: createHash(tx2),
			setupMock: func(mockClient *MockBitcoinClient) {
				mockClient.On("GetRawTransactionVerbose", createHash(tx2)).
					Return(createTxResult(tx2, 0, []btcjson.Vin{
						createVin(tx3),
					}), nil)

				mockClient.On("GetRawTransactionVerbose", createHash(tx3)).
					Return(nil, errors.New("parent rpc error"))
			},
			expectedSize:  1,
			expectedError: "",
		},
		{
			name:   "Success - circular reference prevention",
			txHash: createHash(tx1),
			setupMock: func(mockClient *MockBitcoinClient) {
				// Create a circular reference: tx1 -> tx2 -> tx1
				mockClient.On("GetRawTransactionVerbose", createHash(tx1)).
					Return(createTxResult(tx1, 0, []btcjson.Vin{
						createVin(tx2),
					}), nil).Twice()

				mockClient.On("GetRawTransactionVerbose", createHash(tx2)).
					Return(createTxResult(tx2, 0, []btcjson.Vin{
						createVin(tx1), // Circular reference back to tx1
					}), nil).Twice()

				// Should not call GetRawTransactionVerbose for tx1 again due to visited map
			},
			expectedSize:  2, // Counts tx1 and tx2 before detecting circular reference
			expectedError: "",
		},
		{
			name:   "Error - invalid parent hash",
			txHash: createHash(tx1),
			setupMock: func(mockClient *MockBitcoinClient) {
				mockClient.On("GetRawTransactionVerbose", createHash(tx1)).
					Return(createTxResult(tx1, 0, []btcjson.Vin{
						{Txid: "invalid_hash"}, // This will cause hash parsing to fail
					}), nil)
			},
			expectedSize:  1, // Still counts the current transaction
			expectedError: "",
		},
		{
			name:   "Success - no unconfirmed parents",
			txHash: createHash(tx2),
			setupMock: func(mockClient *MockBitcoinClient) {
				mockClient.On("GetRawTransactionVerbose", createHash(tx2)).
					Return(createTxResult(tx2, 0, []btcjson.Vin{
						createVin(tx3),
						createVin(tx4),
					}), nil)

				// Both parents are confirmed
				mockClient.On("GetRawTransactionVerbose", createHash(tx3)).
					Return(createTxResult(tx3, 3, []btcjson.Vin{}), nil)

				mockClient.On("GetRawTransactionVerbose", createHash(tx4)).
					Return(createTxResult(tx4, 1, []btcjson.Vin{}), nil)
			},
			expectedSize:  1, // Only counts the current transaction
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockBitcoinClient)
			tt.setupMock(mockClient)

			size, err := BtcGetCPFPChainSize(mockClient, tt.txHash)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedSize, size)
			mockClient.AssertExpectations(t)
		})
	}
}
