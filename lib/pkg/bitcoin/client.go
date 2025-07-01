package bitcoin

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
)

const (
	defaultMaxRetries = 3
	defaultRetryDelay = 500 * time.Millisecond
)

type Client struct {
	RPCClient  *rpcclient.Client // Retain for other methods for now. To be removed.
	httpClient *http.Client
	rpcHost    string
	rpcUser    string
	rpcPass    string
}

// jsonRPCRequest defines the structure for a JSON-RPC request.
// We use json.RawMessage for params to allow flexibility for different types.
type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      string          `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

// jsonRPCError defines the structure for a JSON-RPC error object.
type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// jsonRPCResponse defines the structure for a generic JSON-RPC response.
// We use json.RawMessage for result to allow parsing into specific types later.
type jsonRPCResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *jsonRPCError   `json:"error"` // Pointer to allow for null error
	ID     string          `json:"id"`
}

// getBlockHeaderResult is used to unmarshal the relevant part of a getblock response.
type getBlockHeaderResult struct {
	Height int64 `json:"height"`
}

type TxChildrenCount struct {
	ParentTxID    *chainhash.Hash
	ChildrenCount int
}

func NewClient(host string, user string, pass string) (*Client, error) {
	// Legacy RPC Client setup
	connCfg := &rpcclient.ConnConfig{
		Host:         host,
		User:         user,
		Pass:         pass,
		HTTPPostMode: true,
		DisableTLS:   true, // This was the original default
	}

	legacyRPCClient, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create legacy rpc client: %w", err)
	}

	// HTTP client setup for direct sendRequest calls
	url := host
	if strings.HasPrefix(url, "https://") {
		url = "http://" + strings.TrimPrefix(url, "https://")
	} else if !strings.HasPrefix(url, "http://") {
		url = "http://" + url
	}

	// Configure custom transport for better connection pooling
	customTransport := &http.Transport{
		MaxIdleConns:        100,              // Default is 100
		MaxIdleConnsPerHost: 100,              // Increased from default 2
		IdleConnTimeout:     90 * time.Second, // Default is 90 seconds
	}

	httpClient := &http.Client{
		Transport: customTransport,
		Timeout:   30 * time.Second,
	}

	return &Client{
		RPCClient:  legacyRPCClient, // Assigning restored legacy client
		httpClient: httpClient,
		rpcHost:    url,
		rpcUser:    user,
		rpcPass:    pass,
	}, nil
}

func (c *Client) GetBlockHeightByHash(hash *chainhash.Hash) (int64, error) {
	params := []interface{}{hash.String(), 1} // Verbosity 1 for block header
	rawResult, err := c.sendRequest("getblock", params)
	if err != nil {
		return 0, fmt.Errorf("failed to call getblock for hash %s: %w", hash.String(), err)
	}

	var blockResult getBlockHeaderResult
	err = json.Unmarshal(rawResult, &blockResult)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal getblock response for hash %s: %w, body: %s", hash.String(), err, string(rawResult))
	}

	return blockResult.Height, nil
}

// sendRequest is a helper method to make direct JSON-RPC calls.
func (c *Client) sendRequest(method string, params interface{}) (json.RawMessage, error) {
	// Marshal params to json.RawMessage
	paramsData, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params for method %s: %w", method, err)
	}

	reqBody := jsonRPCRequest{
		JSONRPC: "1.0",         // Bitcoin Core typically uses 1.0 for JSON-RPC
		ID:      "gobtcclient", // Can be any string, used for matching responses
		Method:  method,
		Params:  paramsData,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON-RPC request: %w", err)
	}

	var lastErr error

	for attempt := 0; attempt <= defaultMaxRetries; attempt++ {
		req, err := http.NewRequest("POST", c.rpcHost, bytes.NewBuffer(jsonData))
		if err != nil {
			// This error is unlikely to be resolved by retrying, so return immediately.
			return nil, fmt.Errorf("failed to create HTTP request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth(c.rpcUser, c.rpcPass)

		httpResp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("HTTP request failed (attempt %d/%d): %w", attempt+1, defaultMaxRetries+1, err)
			if attempt < defaultMaxRetries {
				time.Sleep(defaultRetryDelay)
				continue
			}
			return nil, lastErr // All retries failed
		}

		respBodyBytes, err := io.ReadAll(httpResp.Body)
		httpResp.Body.Close() // Close body immediately after reading
		if err != nil {
			// Error reading body, less likely to be fixed by retry but could be transient network issue with response stream
			lastErr = fmt.Errorf("failed to read HTTP response body (attempt %d/%d): %w", attempt+1, defaultMaxRetries+1, err)
			if attempt < defaultMaxRetries {
				time.Sleep(defaultRetryDelay)
				continue
			}
			return nil, lastErr // All retries failed
		}

		var rpcResp jsonRPCResponse
		unmarshalErr := json.Unmarshal(respBodyBytes, &rpcResp)

		// Check for HTTP status codes that warrant a retry (e.g., 5xx server errors)
		if httpResp.StatusCode >= 500 && httpResp.StatusCode <= 599 {
			lastErr = fmt.Errorf("HTTP server error (status %s, attempt %d/%d), body: %s", httpResp.Status, attempt+1, defaultMaxRetries+1, string(respBodyBytes))
			if attempt < defaultMaxRetries {
				time.Sleep(defaultRetryDelay)
				continue
			}
			return nil, lastErr // All retries failed for server error
		}

		// Handle non-OK status codes that are not retried (e.g. 4xx)
		if httpResp.StatusCode != http.StatusOK {
			if unmarshalErr == nil && rpcResp.Error != nil {
				return nil, fmt.Errorf("RPC error: %s (code: %d), HTTP status: %s", rpcResp.Error.Message, rpcResp.Error.Code, httpResp.Status)
			}
			return nil, fmt.Errorf("HTTP request failed with status %s, body: %s", httpResp.Status, string(respBodyBytes))
		}

		// If status was OK, but unmarshal failed.
		if unmarshalErr != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON-RPC response: %w, body: %s", unmarshalErr, string(respBodyBytes))
		}

		// If status OK, unmarshal OK, but RPC layer error.
		if rpcResp.Error != nil {
			return nil, fmt.Errorf("RPC error: %s (code: %d)", rpcResp.Error.Message, rpcResp.Error.Code)
		}

		return rpcResp.Result, nil // Success
	}

	return nil, lastErr // Should be unreachable if loop logic is correct, but as a fallback
}

func (c *Client) GetBlockHashByTxID(txID *chainhash.Hash) (*chainhash.Hash, error) {
	// This method now relies on the local GetRawTransactionVerbose, which uses sendRequest.
	tx, err := c.GetRawTransactionVerbose(txID) // Calls the refactored GetRawTransactionVerbose
	if err != nil {
		return nil, fmt.Errorf("failed to get raw transaction verbose for txID %s: %w", txID.String(), err)
	}
	if tx.BlockHash == "" {
		return nil, fmt.Errorf("transaction %s not yet mined or blockhash not available", txID.String())
	}
	return chainhash.NewHashFromStr(tx.BlockHash)
}

func (c *Client) GetTxProof(txID *chainhash.Hash, blockHash *chainhash.Hash) (string, error) {
	params := []interface{}{
		[]string{txID.String()}, // txids must be an array of strings
		blockHash.String(),      // blockhash is a string
	}

	rawResult, err := c.sendRequest("gettxoutproof", params)
	if err != nil {
		return "", fmt.Errorf("failed to call gettxoutproof for tx %s in block %s: %w", txID.String(), blockHash.String(), err)
	}

	var proof string
	err = json.Unmarshal(rawResult, &proof)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal gettxoutproof response: %w, body: %s", err, string(rawResult))
	}

	return proof, nil
}

func (c *Client) GetBlockHashesByStartHeight(startHeight int64, count int64) ([]*chainhash.Hash, error) {
	if count <= 0 {
		return []*chainhash.Hash{}, nil
	}
	blockHashes := make([]*chainhash.Hash, 0, count)

	for i := int64(0); i < count; i++ {
		currentHeight := startHeight + i
		params := []interface{}{currentHeight}

		rawResult, err := c.sendRequest("getblockhash", params)
		if err != nil {
			return nil, fmt.Errorf("error fetching block hash for height %d: %w", currentHeight, err)
		}

		var hashStr string
		if err := json.Unmarshal(rawResult, &hashStr); err != nil {
			return nil, fmt.Errorf("error unmarshaling block hash string for height %d: %w, raw: %s", currentHeight, err, string(rawResult))
		}

		blockHash, err := chainhash.NewHashFromStr(hashStr)
		if err != nil {
			return nil, fmt.Errorf("error creating chainhash.Hash from string for height %d: %w, hashStr: %s", currentHeight, err, hashStr)
		}
		blockHashes = append(blockHashes, blockHash)
	}

	return blockHashes, nil
}

func (c *Client) GetRawTransactionVerbose(txHash *chainhash.Hash) (*btcjson.TxRawResult, error) {
	// Fallback to manual implementation if RPCClient is not available
	params := []interface{}{txHash.String(), true}
	rawResult, err := c.sendRequest("getrawtransaction", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get raw transaction: %w", err)
	}

	var txResult btcjson.TxRawResult
	err = json.Unmarshal(rawResult, &txResult)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction result: %w", err)
	}

	return &txResult, nil
}

func (c *Client) GetTxChildrenCount(parentHash *chainhash.Hash) (*TxChildrenCount, error) {
	// 1. Verify parent exists first
	_, err := c.RPCClient.GetRawTransaction(parentHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get parent transaction: %v", err)
	}

	// 2. Get mempool (string IDs)
	mempool, err := c.RPCClient.GetRawMempool()
	if err != nil {
		return nil, fmt.Errorf("failed to get mempool: %v", err)
	}

	// 3. Batch get all mempool transactions
	type job struct {
		txID string
		idx  int
	}
	type result struct {
		inputs int
		err    error
	}

	workers := 8 // Tune based on your RPC server capacity
	jobChan := make(chan job, len(mempool))
	resultChan := make(chan result, len(mempool))

	// Worker pool
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobChan {
				txHash, err := chainhash.NewHashFromStr(j.txID)
				if err != nil {
					resultChan <- result{err: fmt.Errorf("failed to parse txID %s: %v", j.txID, err)}
					continue
				}

				tx, err := c.RPCClient.GetRawTransaction(txHash)
				if err != nil {
					resultChan <- result{err: fmt.Errorf("failed to get tx %s: %v", txHash, err)}
					continue
				}

				inputs := 0
				for _, txIn := range tx.MsgTx().TxIn {
					if txIn.PreviousOutPoint.Hash.IsEqual(parentHash) {
						inputs++
					}
				}
				resultChan <- result{inputs: inputs}
			}
		}()
	}

	// Feed jobs
	for i, txID := range mempool {
		jobChan <- job{txID: txID.String(), idx: i}
	}
	close(jobChan)
	wg.Wait()
	close(resultChan)

	// Process results
	childrenCount := 0
	for res := range resultChan {
		fmt.Printf("tx %d inputs\n", res.inputs)
		if res.err != nil {
			logger.Log.Error().Err(res.err).Msg("tx processing error")
			continue
		}
		if res.inputs > 0 {
			childrenCount++
		}
	}

	return &TxChildrenCount{
		ParentTxID:    parentHash,
		ChildrenCount: childrenCount,
	}, nil
}

func (c *Client) SendRawTransaction(tx *wire.MsgTx, allowHighFees bool) (*chainhash.Hash, error) {
	// Fallback to manual implementation if RPCClient is not available
	// Serialize the transaction to hex
	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}
	hexTx := hex.EncodeToString(buf.Bytes())

	params := []interface{}{hexTx}
	if allowHighFees {
		params = append(params, allowHighFees)
	}

	rawResult, err := c.sendRequest("sendrawtransaction", params)
	if err != nil {
		return nil, fmt.Errorf("failed to send raw transaction: %w", err)
	}

	var txIDStr string
	err = json.Unmarshal(rawResult, &txIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction ID: %w", err)
	}

	return chainhash.NewHashFromStr(txIDStr)
}

func (c *Client) GetBlockChainInfo() (*btcjson.GetBlockChainInfoResult, error) {
	rawResult, err := c.sendRequest("getblockchaininfo", nil) // No parameters for getblockchaininfo
	if err != nil {
		return nil, fmt.Errorf("failed to call getblockchaininfo: %w", err)
	}

	var chainInfoResult btcjson.GetBlockChainInfoResult
	err = json.Unmarshal(rawResult, &chainInfoResult)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal getblockchaininfo result: %w, body: %s", err, string(rawResult))
	}

	return &chainInfoResult, nil
}

// GetTxOut retrieves details about an unspent transaction output.
// It takes a transaction ID (txid), the output index number (vout), and a boolean
// indicating whether to include the mempool (includeMempool).
func (c *Client) GetTxOut(txid *chainhash.Hash, vout uint32, includeMempool bool) (*btcjson.GetTxOutResult, error) {
	params := []interface{}{txid.String(), vout, includeMempool}
	rawResult, err := c.sendRequest("gettxout", params)
	if err != nil {
		// Bitcoin Core returns a null result and an error code of -5 if the UTXO is spent or not found.
		// We can check for this specific error if needed, but for now, just forwarding the error.
		// Example of a more specific error check:
		// if rpcErr, ok := err.(*btcjson.RPCError); ok && rpcErr.Code == btcjson.ErrRPCVerifyAlreadyInChain {
		// 	 return nil, nil // Or a specific error indicating UTXO not found/spent
		// }
		return nil, fmt.Errorf("failed to call gettxout for tx %s, vout %d: %w", txid.String(), vout, err)
	}

	// We need to handle this case. If rawResult is nil OR if it represents the JSON value 'null', it means the UTXO is spent.
	if rawResult == nil || string(rawResult) == "null" {
		return nil, nil // Indicates UTXO is spent or not found
	}

	var txOutResult btcjson.GetTxOutResult
	err = json.Unmarshal(rawResult, &txOutResult)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal gettxout result for tx %s, vout %d: %w, body: %s", txid.String(), vout, err, string(rawResult))
	}

	return &txOutResult, nil
}

func (c *Client) ShutdownRPCClient() {
	// Shutdown the legacy client if it's used (restored)
	if c.RPCClient != nil {
		c.RPCClient.Shutdown()
	}

	// Close idle connections for the direct http client (remains as is)
	if c.httpClient != nil && c.httpClient.Transport != nil {
		if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}
}
