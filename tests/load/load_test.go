package load

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

func TestMain(m *testing.M) {
	redisClient = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	code := m.Run()

	redisClient.Close()
	os.Exit(code)
}
func TestHashStrategy(t *testing.T) {
	t.Log("Testing Hash-based Token Bucket Strategy")
	runStrategyTests(t, "tokenbucket-hash")
}
func TestTransactionStrategy(t *testing.T) {
	t.Log("Testing Transaction-based Token Bucket Strategy")
	runStrategyTests(t, "tokenbucket-transaction")
}
func TestLuaStrategy(t *testing.T) {
	t.Log("Testing Lua Script-based Token Bucket Strategy")
	runStrategyTests(t, "tokenbucket-lua")
}
func runStrategyTests(t *testing.T, strategy string) {
	t.Helper()

	scenarios := []struct {
		name string
		fn   func(*testing.T, string) TestResult
	}{
		{"Single Client Burst", testSingleClientBurst},
		{"Multiple Clients Concurrent", testMultipleClientsConcurrent},
		{"Burst Then Steady", testBurstThenSteady},
	}

	results := make([]TestResult, 0, len(scenarios))

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Cleanup Redis before each test
			cleanupRedis(t)
			time.Sleep(100 * time.Millisecond)

			result := scenario.fn(t, strategy)
			result.Strategy = strategy
			result.Scenario = scenario.name
			results = append(results, result)

			printTestResult(t, result)
		})
	}

	t.Run("Summary", func(t *testing.T) {
		generateSummaryReport(t, strategy, results)
	})
}
func cleanupRedis(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	iter := redisClient.Scan(ctx, 0, "ratelimit:*", 0).Iterator()
	for iter.Next(ctx) {
		if err := redisClient.Del(ctx, iter.Val()).Err(); err != nil {
			t.Logf("Warning: failed to delete key %s: %v", iter.Val(), err)
		}
	}
	if err := iter.Err(); err != nil {
		t.Logf("Warning: Redis scan error: %v", err)
	}
}
func printTestResult(t *testing.T, result TestResult) {
	t.Helper()

	t.Logf("\n"+
		"========================================\n"+
		"Strategy: %s\n"+
		"Scenario: %s\n"+
		"========================================\n"+
		"Total Requests: %d\n"+
		"Successful (200): %d\n"+
		"Rate Limited (429): %d\n"+
		"Errors: %d\n"+
		"Duration: %v\n"+
		"Throughput: %.2f req/sec\n"+
		"----------------------------------------\n"+
		"Latency p50: %v\n"+
		"Latency p95: %v\n"+
		"Latency p99: %v\n"+
		"Latency max: %v\n"+
		"----------------------------------------\n"+
		"Race Condition Check:\n"+
		"Expected Max Success: %d\n"+
		"Actual Success: %d\n"+
		"Over-limit Grants: %d\n"+
		"Race Detected: %v\n"+
		"========================================\n",
		result.Strategy,
		result.Scenario,
		result.TotalRequests,
		result.SuccessCount,
		result.RateLimitCount,
		result.ErrorCount,
		result.Duration,
		result.Throughput,
		result.Latency.P50,
		result.Latency.P95,
		result.Latency.P99,
		result.Latency.Max,
		result.ExpectedMaxSuccess,
		result.SuccessCount,
		result.OverLimitGrants,
		result.RaceDetected,
	)
}
func generateSummaryReport(t *testing.T, strategy string, results []TestResult) {
	t.Helper()

	t.Logf("\n"+
		"########################################\n"+
		"# STRATEGY SUMMARY: %s\n"+
		"########################################\n",
		strategy,
	)

	for _, result := range results {
		t.Logf("%-30s | Throughput: %6.2f req/s | p95: %6s | Races: %v\n",
			result.Scenario,
			result.Throughput,
			result.Latency.P95,
			result.RaceDetected,
		)
	}

	t.Log("########################################\n")
}
