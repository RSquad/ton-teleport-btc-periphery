package bitcoin

import (
	"encoding/hex"
	"sync"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/joho/godotenv"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testBitcoinConfig struct {
	Host string `env:"COMMON_BITCOIN_RPC_HOST,required"`
	User string `env:"COMMON_BITCOIN_RPC_USER,required"`
	Pass string `env:"COMMON_BITCOIN_RPC_PASS,required"`
}

func setupMainTestClient(t *testing.T) *Client {
	t.Helper()
	dotEnvPath := "../../../.env"
	if err := godotenv.Load(dotEnvPath); err != nil {
		t.Logf("Could not load .env file from %s: %v. Will proceed and rely on utils.LoadCfg or existing environment variables.", dotEnvPath, err)
	}

	cfg, err := utils.LoadCfg[testBitcoinConfig]()
	if err != nil {
		t.Skipf("Skipping Bitcoin client test: Failed to load config: %v. Ensure RPC vars are set.", err)
	}
	if cfg.Host == "" || cfg.User == "" || cfg.Pass == "" {
		t.Skip("Skipping Bitcoin client test: Config loaded but RPC vars are empty.")
	}

	client, err := NewClient(cfg.Host, cfg.User, cfg.Pass)
	require.NoError(t, err, "Failed to create Bitcoin client")
	require.NotNil(t, client, "Bitcoin client should not be nil")
	require.NotNil(t, client.RPCClient, "client.RPCClient (legacy client) should not be nil")
	return client
}

type testPrerequisites struct {
	blockHash   *chainhash.Hash
	blockHeight int64
	txHash      *chainhash.Hash
	tx          *btcjson.TxRawResult
	rawBlock    *btcjson.GetBlockVerboseResult
}

func getTestPrerequisites(t *testing.T, client *Client) testPrerequisites {
	t.Helper()

	info, err := client.RPCClient.GetBlockChainInfo()
	require.NoError(t, err)
	require.NotNil(t, info)

	const blocksToStepBack = 5
	targetHeight := info.Blocks - blocksToStepBack
	if info.Blocks <= blocksToStepBack && info.Blocks > 0 {
		targetHeight = info.Blocks
	} else if info.Blocks == 0 {
		t.Skip("Skipping test: blockchain has no blocks.")
	}

	blockHash, err := client.RPCClient.GetBlockHash(int64(targetHeight))
	require.NoError(t, err, "Failed to get block hash for height %d", targetHeight)
	require.NotNil(t, blockHash)

	blockVerbose, err := client.RPCClient.GetBlockVerbose(blockHash)
	require.NoError(t, err, "Failed to get block verbose for hash %s", blockHash.String())
	require.NotNil(t, blockVerbose)
	require.NotEmpty(t, blockVerbose.Tx, "Selected block %s has no transactions", blockHash.String())

	var selectedTxHashStr string
	if len(blockVerbose.Tx) > 1 {
		selectedTxHashStr = blockVerbose.Tx[1]
	} else {
		selectedTxHashStr = blockVerbose.Tx[0]
	}

	txHash, err := chainhash.NewHashFromStr(selectedTxHashStr)
	require.NoError(t, err, "Failed to create chainhash.Hash from tx string %s", selectedTxHashStr)

	txDetails, err := client.RPCClient.GetRawTransactionVerbose(txHash)
	require.NoError(t, err, "Failed to get raw transaction verbose for tx %s", txHash.String())
	require.NotNil(t, txDetails)
	require.NotEmpty(t, txDetails.BlockHash, "Transaction %s used for prerequisites is not confirmed", txHash.String())
	require.Equal(t, blockHash.String(), txDetails.BlockHash, "Transaction %s is not in expected block %s", txHash.String(), blockHash.String())

	return testPrerequisites{
		blockHash:   blockHash,
		blockHeight: blockVerbose.Height,
		txHash:      txHash,
		tx:          txDetails,
		rawBlock:    blockVerbose,
	}
}

func TestClient_CompareConcurrentVsSequentialForGetBlockChainInfo(t *testing.T) {
	const (
		numRequestsConst = 50
	)

	client := setupMainTestClient(t)
	defer client.ShutdownRPCClient()

	t.Logf("Starting %d CONCURRENT GetBlockChainInfo requests...", numRequestsConst)
	var wg sync.WaitGroup
	wg.Add(numRequestsConst)

	concurrentResults := make(map[int]*btcjson.GetBlockChainInfoResult)
	var concurrentResultsMutex sync.Mutex

	concurrentStartTime := time.Now()

	for i := 0; i < numRequestsConst; i++ {
		go func(idx int) {
			defer wg.Done()
			info, errGR := client.GetBlockChainInfo()

			if errGR != nil {
				t.Errorf("Goroutine %d: Failed to get blockchain info: %v", idx, errGR)
				return
			}
			if info == nil {
				t.Errorf("Goroutine %d: Blockchain info result should not be nil", idx)
				return
			}

			concurrentResultsMutex.Lock()
			concurrentResults[idx] = info
			concurrentResultsMutex.Unlock()
		}(i)
	}

	wg.Wait()
	concurrentEndTime := time.Now()
	totalConcurrentDuration := concurrentEndTime.Sub(concurrentStartTime)
	t.Logf("Completed %d CONCURRENT requests. Total time: %v", numRequestsConst, totalConcurrentDuration)

	t.Logf("Starting %d SEQUENTIAL GetBlockChainInfo requests...", numRequestsConst)
	sequentialResults := make([]*btcjson.GetBlockChainInfoResult, numRequestsConst)

	sequentialStartTime := time.Now()
	for i := 0; i < numRequestsConst; i++ {
		info, errSeq := client.GetBlockChainInfo()
		require.NoError(t, errSeq, "Sequential request %d: Failed to get blockchain info", i)
		require.NotNil(t, info, "Sequential request %d: Blockchain info result should not be nil", i)
		sequentialResults[i] = info
	}
	sequentialEndTime := time.Now()
	totalSequentialDuration := sequentialEndTime.Sub(sequentialStartTime)
	t.Logf("Completed %d SEQUENTIAL requests. Total time: %v", numRequestsConst, totalSequentialDuration)

	t.Logf("Comparing results from concurrent and sequential executions...")
	if len(concurrentResults) != numRequestsConst {
		t.Fatalf("Expected %d successful concurrent results, but got %d. Check t.Errorf messages above.", numRequestsConst, len(concurrentResults))
	}
	require.Equal(t, numRequestsConst, len(sequentialResults), "Number of results from sequential execution does not match numRequests")

	for i := 0; i < numRequestsConst; i++ {
		concurrentInfo, ok := concurrentResults[i]
		require.True(t, ok, "Blockchain info for index %d not found in concurrent results map", i)
		sequentialInfo := sequentialResults[i]
		assert.Equal(t, sequentialInfo.Chain, concurrentInfo.Chain, "Chain mismatch for index %d", i)
		assert.Equal(t, sequentialInfo.Blocks, concurrentInfo.Blocks, "Blocks mismatch for index %d", i)
		assert.Equal(t, sequentialInfo.Headers, concurrentInfo.Headers, "Headers mismatch for index %d", i)
	}
	t.Logf("All %d blockchain info details from concurrent and sequential executions match.", numRequestsConst)

	t.Logf("Concurrent duration: %v, Sequential duration: %v", totalConcurrentDuration, totalSequentialDuration)

	thresholdFactor := 0.50
	maxAllowedConcurrentDuration := time.Duration(float64(totalSequentialDuration) * thresholdFactor)

	if totalSequentialDuration < 100*time.Millisecond {
		maxAllowedConcurrentDuration = totalSequentialDuration
		t.Logf("Sequential execution was very fast (%v), adjusting speed comparison threshold.", totalSequentialDuration)
	}

	condition := totalConcurrentDuration < maxAllowedConcurrentDuration
	if totalSequentialDuration == 0 && totalConcurrentDuration == 0 {
		condition = true
	}

	assert.True(t, condition,
		"Concurrent execution (%v) was not significantly faster than sequential execution (%v). Max allowed concurrent: %v",
		totalConcurrentDuration, totalSequentialDuration, maxAllowedConcurrentDuration)

	if condition {
		t.Logf("Concurrent execution was faster or acceptably close for very fast sequential, and results match.")
	} else {
		failureRate := float64(totalConcurrentDuration-maxAllowedConcurrentDuration) / float64(maxAllowedConcurrentDuration) * 100
		if maxAllowedConcurrentDuration == 0 && totalConcurrentDuration > 0 {
			t.Logf("Concurrent execution (%v) was slower than sequential execution (%v), which was instantaneous.", totalConcurrentDuration, totalSequentialDuration)
		} else if maxAllowedConcurrentDuration == 0 && totalConcurrentDuration == 0 {
		} else {
			t.Logf("Concurrent execution missed the speed target by %.2f%%.", failureRate)
		}
	}
}

func TestClient_GetBlockChainInfo_Comparison(t *testing.T) {
	client := setupMainTestClient(t)
	defer client.ShutdownRPCClient()

	myInfo, myErr := client.GetBlockChainInfo()
	require.NoError(t, myErr)
	require.NotNil(t, myInfo)

	legacyInfo, legacyErr := client.RPCClient.GetBlockChainInfo()
	require.NoError(t, legacyErr)
	require.NotNil(t, legacyInfo)

	assert.Equal(t, legacyInfo.Chain, myInfo.Chain)
	assert.True(t, myInfo.Blocks > 0 || legacyInfo.Blocks == 0)
	assert.True(t, legacyInfo.Blocks > 0 || legacyInfo.Blocks == 0)
	assert.NotEmpty(t, myInfo.BestBlockHash)
	assert.NotEmpty(t, legacyInfo.BestBlockHash)

	assert.Equal(t, legacyInfo.InitialBlockDownload, myInfo.InitialBlockDownload)
	assert.Equal(t, legacyInfo.Pruned, myInfo.Pruned)
}

func TestClient_GetBlockHeightByHash(t *testing.T) {
	client := setupMainTestClient(t)
	defer client.ShutdownRPCClient()

	prereqs := getTestPrerequisites(t, client)

	myHeight, myErr := client.GetBlockHeightByHash(prereqs.blockHash)
	require.NoError(t, myErr)

	_, legacyErr := client.RPCClient.GetBlockHeader(prereqs.blockHash)
	require.NoError(t, legacyErr, "Legacy GetBlockHeader failed")

	assert.Equal(t, prereqs.blockHeight, myHeight, "Block height does not match prerequisite block height")
}

func TestClient_GetBlockHashByTxID(t *testing.T) {
	client := setupMainTestClient(t)
	defer client.ShutdownRPCClient()

	prereqs := getTestPrerequisites(t, client)
	require.NotEmpty(t, prereqs.tx.BlockHash, "Test transaction must be confirmed")

	myBlockHash, myErr := client.GetBlockHashByTxID(prereqs.txHash)
	require.NoError(t, myErr)
	require.NotNil(t, myBlockHash)

	legacyTxInfo, legacyErr := client.RPCClient.GetRawTransactionVerbose(prereqs.txHash)
	require.NoError(t, legacyErr)
	require.NotEmpty(t, legacyTxInfo.BlockHash)
	legacyBlockHash, err := chainhash.NewHashFromStr(legacyTxInfo.BlockHash)
	require.NoError(t, err)

	assert.Equal(t, legacyBlockHash, myBlockHash, "Block hashes for TXID do not match")
	assert.Equal(t, prereqs.blockHash, myBlockHash, "Block hash for TXID does not match prerequisite block hash")
}

func TestClient_GetTxProof(t *testing.T) {
	client := setupMainTestClient(t)
	defer client.ShutdownRPCClient()

	prereqs := getTestPrerequisites(t, client)
	require.NotEmpty(t, prereqs.tx.BlockHash, "Test transaction must be confirmed for GetTxProof")
	require.Equal(t, prereqs.blockHash.String(), prereqs.tx.BlockHash, "Tx block hash mismatch, prerequisite data is inconsistent.")

	// Sanity check with legacy client that the transaction is indeed in the block specified by prereqs.
	// This is crucial because gettxoutproof relies on this fact.
	legacyTxVerbose, err := client.RPCClient.GetRawTransactionVerbose(prereqs.txHash)
	require.NoError(t, err, "Failed to get legacy tx verbose for sanity check")
	require.Equal(t, prereqs.blockHash.String(), legacyTxVerbose.BlockHash,
		"Sanity check failed: Legacy client reports tx %s is in block %s, expected %s",
		prereqs.txHash.String(), legacyTxVerbose.BlockHash, prereqs.blockHash.String())

	myProofHex, myErr := client.GetTxProof(prereqs.txHash, prereqs.blockHash)
	require.NoError(t, myErr, "client.GetTxProof failed for tx %s in block %s", prereqs.txHash, prereqs.blockHash)
	require.NotEmpty(t, myProofHex, "client.GetTxProof returned an empty proof string for tx %s in block %s", prereqs.txHash, prereqs.blockHash)

	// Basic validation of the hex proof string from the custom client.
	_, err = hex.DecodeString(myProofHex)
	assert.NoError(t, err, "Proof string from client.GetTxProof ('%s') is not valid hex", myProofHex)
}

func TestClient_GetBlockHashesByStartHeight(t *testing.T) {
	client := setupMainTestClient(t)
	defer client.ShutdownRPCClient()

	prereqs := getTestPrerequisites(t, client)

	count := int64(3)
	startHeight := prereqs.blockHeight - count + 1
	if startHeight < 0 {
		startHeight = 0
	}
	if startHeight == 0 && count == 0 {
		myHashes, myErr := client.GetBlockHashesByStartHeight(startHeight, count)
		require.NoError(t, myErr)
		require.Empty(t, myHashes)
		return
	}
	if count == 0 {
		myHashes, myErr := client.GetBlockHashesByStartHeight(startHeight, count)
		require.NoError(t, myErr)
		require.Empty(t, myHashes)
		return
	}

	chainInfo, err := client.RPCClient.GetBlockChainInfo()
	require.NoError(t, err)
	if startHeight+count-1 > int64(chainInfo.Blocks) {
		t.Skipf("Skipping test: requested block range %d-%d (count %d) exceeds current chain height %d", startHeight, startHeight+count-1, count, chainInfo.Blocks)
		return
	}
	if startHeight < 0 {
		t.Skipf("Skipping test: startHeight %d is invalid for current chain height %d", startHeight, chainInfo.Blocks)
		return
	}

	myHashes, myErr := client.GetBlockHashesByStartHeight(startHeight, count)
	require.NoError(t, myErr)
	require.Len(t, myHashes, int(count))

	legacyHashes := make([]*chainhash.Hash, 0, count)
	for i := int64(0); i < count; i++ {
		currentTestHeight := startHeight + i
		if currentTestHeight > int64(chainInfo.Blocks) {
			t.Logf("Adjusting test: currentTestHeight %d exceeds chainInfo.Blocks %d. Shortening comparison.", currentTestHeight, chainInfo.Blocks)
			break
		}
		h, err := client.RPCClient.GetBlockHash(currentTestHeight)
		require.NoError(t, err, "Failed to get legacy block hash for height %d", currentTestHeight)
		legacyHashes = append(legacyHashes, h)
	}
	require.Len(t, legacyHashes, len(myHashes), "Mismatch in number of hashes fetched if test range was adjusted")

	assert.Equal(t, legacyHashes, myHashes, "Block hash sequences do not match")
}

func TestClient_GetRawTransactionVerbose(t *testing.T) {
	client := setupMainTestClient(t)
	defer client.ShutdownRPCClient()

	prereqs := getTestPrerequisites(t, client)

	myTxInfo, myErr := client.GetRawTransactionVerbose(prereqs.txHash)
	require.NoError(t, myErr)
	require.NotNil(t, myTxInfo)

	legacyTxInfo, legacyErr := client.RPCClient.GetRawTransactionVerbose(prereqs.txHash)
	require.NoError(t, legacyErr)
	require.NotNil(t, legacyTxInfo)

	assert.Equal(t, legacyTxInfo.Txid, myTxInfo.Txid)
	assert.Equal(t, legacyTxInfo.Hash, myTxInfo.Hash)
	assert.Equal(t, legacyTxInfo.Version, myTxInfo.Version)
	assert.Equal(t, legacyTxInfo.LockTime, myTxInfo.LockTime)
	assert.Equal(t, legacyTxInfo.BlockHash, myTxInfo.BlockHash)
	if legacyTxInfo.BlockHash != "" {
		assert.True(t, myTxInfo.Confirmations >= legacyTxInfo.Confirmations || myTxInfo.Confirmations >= prereqs.tx.Confirmations-1, "Confirmations mismatch: my=%d, legacy=%d, prereq_tx_conf=%d", myTxInfo.Confirmations, legacyTxInfo.Confirmations, prereqs.tx.Confirmations)
	} else {
		assert.Equal(t, int64(0), myTxInfo.Confirmations)
	}
	assert.Equal(t, legacyTxInfo.Hex, myTxInfo.Hex)
	assert.Equal(t, len(legacyTxInfo.Vin), len(myTxInfo.Vin))
	assert.Equal(t, len(legacyTxInfo.Vout), len(myTxInfo.Vout))
}

func TestClient_GetTxOut(t *testing.T) {
	client := setupMainTestClient(t)
	defer client.ShutdownRPCClient()

	prereqs := getTestPrerequisites(t, client)
	require.NotEmpty(t, prereqs.tx.Vout, "Transaction for GetTxOut test has no outputs")

	voutIndex := prereqs.tx.Vout[0].N
	includeMempool := true

	myTxOut, myErr := client.GetTxOut(prereqs.txHash, voutIndex, includeMempool)
	legacyTxOut, legacyErr := client.RPCClient.GetTxOut(prereqs.txHash, voutIndex, includeMempool)

	require.Equal(t, legacyErr, myErr, "Error status mismatch between client and legacy client for existing vout")

	if legacyTxOut == nil {
		assert.Nil(t, myTxOut, "My client should also return nil for a spent/non-existent UTXO when legacy does")
	} else {
		require.NotNil(t, myTxOut, "My client returned nil for an existing UTXO when legacy found one")
		assert.Equal(t, legacyTxOut.BestBlock, myTxOut.BestBlock)
		if legacyTxOut.Confirmations > 0 {
			assert.True(t, myTxOut.Confirmations > 0, "My UTXO shows 0 confirmations when legacy shows %d", legacyTxOut.Confirmations)
			assert.InDelta(t, legacyTxOut.Confirmations, myTxOut.Confirmations, 2, "Confirmations differ by more than 2: legacy=%d, my=%d", legacyTxOut.Confirmations, myTxOut.Confirmations)
		} else {
			assert.Equal(t, int64(0), myTxOut.Confirmations, "My UTXO shows confirmations when legacy shows 0 (mempool)")
		}
		assert.Equal(t, legacyTxOut.Value, myTxOut.Value)
		assert.Equal(t, legacyTxOut.ScriptPubKey.Hex, myTxOut.ScriptPubKey.Hex)
		assert.Equal(t, legacyTxOut.ScriptPubKey.Type, myTxOut.ScriptPubKey.Type)
		assert.Equal(t, legacyTxOut.Coinbase, myTxOut.Coinbase)
	}

	nonExistentVoutIndex := uint32(99999)
	myTxOutNE, myErrNE := client.GetTxOut(prereqs.txHash, nonExistentVoutIndex, includeMempool)
	legacyTxOutNE, legacyErrNE := client.RPCClient.GetTxOut(prereqs.txHash, nonExistentVoutIndex, includeMempool)

	assert.Nil(t, myErrNE, "My client GetTxOut returned an error for a non-existent vout: %v", myErrNE)
	assert.Nil(t, legacyErrNE, "Legacy client GetTxOut returned an error for a non-existent vout: %v", legacyErrNE)
	assert.Nil(t, myTxOutNE, "My client should return nil data for a non-existent Vout")
	assert.Nil(t, legacyTxOutNE, "Legacy client should return nil data for a non-existent Vout")
}

func TestClient_SendRawTransaction(t *testing.T) {
	client := setupMainTestClient(t)
	defer client.ShutdownRPCClient()

	info, err := client.RPCClient.GetBlockChainInfo()
	require.NoError(t, err)
	if info.Chain != "regtest" {
		t.Skip("Skipping SendRawTransaction test: not on regtest. Current chain: ", info.Chain)
	}

	tx := wire.NewMsgTx(wire.TxVersion)
	prevTxHash, _ := chainhash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000001")
	outPoint := wire.NewOutPoint(prevTxHash, 0)
	txIn := wire.NewTxIn(outPoint, []byte{txscript.OP_0, txscript.OP_0}, nil)
	tx.AddTxIn(txIn)

	scriptPubKey, err := hex.DecodeString("76a914000000000000000000000000000000000000000088ac")
	require.NoError(t, err)
	txOut := wire.NewTxOut(0, scriptPubKey)
	tx.AddTxOut(txOut)

	txHash, err := client.SendRawTransaction(tx, false)
	if err != nil {
		t.Logf("Successfully received expected error from SendRawTransaction for invalid dummy tx: %v", err)
		assert.Error(t, err)
		assert.Nil(t, txHash, "TxHash should be nil on error")
	} else {
		t.Errorf("SendRawTransaction did not return an error for an obviously invalid tx. TxHash: %s", txHash.String())
		t.Logf("Unexpected success. This might indicate the regtest node accepted a strange tx or the test logic needs review.")
	}
}
