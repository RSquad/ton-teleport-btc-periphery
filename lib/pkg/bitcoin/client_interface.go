package bitcoin

import (
	"net/http"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
)

// ClientInterface defines all the methods that a Bitcoin client should implement
type ClientInterface interface {
	GetBlockChainInfo() (*btcjson.GetBlockChainInfoResult, error)
	GetRawTransactionVerbose(txHash *chainhash.Hash) (*btcjson.TxRawResult, error)
	GetBlockHeightByHash(hash *chainhash.Hash) (int64, error)
	GetBlockHashByTxID(txID *chainhash.Hash) (*chainhash.Hash, error)
	GetTxProof(txID *chainhash.Hash, blockHash *chainhash.Hash) (string, error)
	SendRawTransaction(tx *wire.MsgTx, allowHighFees bool) (*chainhash.Hash, error)
	GetTxOut(txid *chainhash.Hash, vout uint32, includeMempool bool) (*btcjson.GetTxOutResult, error)
	GetBlockHashesByStartHeight(startHeight int64, count int64) ([]*chainhash.Hash, error)
	GetTxChildrenCount(parentHash *chainhash.Hash) (*TxChildrenCount, error)
	EstimateFee(blockCount int, estimateMode *btcjson.EstimateSmartFeeMode) (float64, error)
	GetRPCClient() *rpcclient.Client
	GetHTTPClient() *http.Client
	ShutdownRPCClient()
}

// Ensure Client implements ClientInterface
var _ ClientInterface = (*Client)(nil)
