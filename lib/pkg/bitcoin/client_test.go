package bitcoin

import (
	"sync"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcjson"
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

func TestClient_CompareConcurrentVsSequentialForGetBlockChainInfo(t *testing.T) {
	// Test-specific constants
	const (
		numRequestsConst = 50 // Number of times to fetch the same information
	)

	dotEnvPath := "../../../.env" // Adjusted path if tests are run from package dir
	err := godotenv.Load(dotEnvPath)
	if err != nil {
		t.Logf("Could not load .env file from %s: %v. Will proceed and rely on utils.LoadCfg or existing environment variables.", dotEnvPath, err)
	}

	cfg, err := utils.LoadCfg[testBitcoinConfig]()
	if err != nil {
		t.Skipf("Skipping Bitcoin client test: Failed to load config: %v. Ensure RPC vars are set.", err)
		return
	}

	if cfg.Host == "" || cfg.User == "" || cfg.Pass == "" {
		t.Skip("Skipping Bitcoin client test: Config loaded but RPC vars are empty.")
		return
	}

	client, err := NewClient(cfg.Host, cfg.User, cfg.Pass)
	require.NoError(t, err, "Failed to create Bitcoin client")
	require.NotNil(t, client, "Bitcoin client should not be nil")
	defer client.ShutdownRPCClient()

	// --- Concurrent Execution ---
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

			// Use assert instead of require in goroutines to allow other goroutines to complete
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

	// --- Sequential Execution ---
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

	// --- Result Comparison ---
	t.Logf("Comparing results from concurrent and sequential executions...")
	// Check if all concurrent requests were successful before comparing lengths
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
		// Optionally, compare more fields if necessary
	}
	t.Logf("All %d blockchain info details from concurrent and sequential executions match.", numRequestsConst)

	// --- Speed Comparison ---
	t.Logf("Concurrent duration: %v, Sequential duration: %v", totalConcurrentDuration, totalSequentialDuration)

	thresholdFactor := 0.50 // Expecting at least 50% speedup
	maxAllowedConcurrentDuration := time.Duration(float64(totalSequentialDuration) * thresholdFactor)

	// Allow concurrent to be slightly slower if sequential is very fast (e.g. < 100ms for all requests)
	// to avoid flakiness due to goroutine overhead on very quick operations.
	if totalSequentialDuration < 100*time.Millisecond {
		maxAllowedConcurrentDuration = totalSequentialDuration // Must be faster or equal for very quick ops
		t.Logf("Sequential execution was very fast (%v), adjusting speed comparison threshold.", totalSequentialDuration)
	}

	condition := totalConcurrentDuration < maxAllowedConcurrentDuration
	if totalSequentialDuration == 0 && totalConcurrentDuration == 0 { // Avoid division by zero or issues if times are identical and zero
		condition = true // If both are zero, consider it a pass for speed
	}

	assert.True(t, condition,
		"Concurrent execution (%v) was not significantly faster than sequential execution (%v). Max allowed concurrent: %v",
		totalConcurrentDuration, totalSequentialDuration, maxAllowedConcurrentDuration)

	if condition {
		t.Logf("Concurrent execution was faster or acceptably close for very fast sequential, and results match.")
	} else {
		failureRate := float64(totalConcurrentDuration-maxAllowedConcurrentDuration) / float64(maxAllowedConcurrentDuration) * 100
		t.Logf("Concurrent execution missed the speed target by %.2f%%.", failureRate)
	}
}
