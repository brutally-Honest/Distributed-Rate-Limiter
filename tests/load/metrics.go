package load

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type TestResult struct {
	Strategy           string
	Scenario           string
	TotalRequests      int
	SuccessCount       int
	RateLimitCount     int
	ErrorCount         int
	Duration           time.Duration
	Throughput         float64
	Latency            LatencyMetrics
	ExpectedMaxSuccess int
	OverLimitGrants    int
	RaceDetected       bool
}

// LatencyMetrics holds percentile latency data
type LatencyMetrics struct {
	P50 time.Duration
	P95 time.Duration
	P99 time.Duration
	Max time.Duration
	Min time.Duration
	Avg time.Duration
}

// MetricsCollector collects metrics concurrently
type MetricsCollector struct {
	mu                 sync.Mutex
	successCount       atomic.Int64
	rateLimitCount     atomic.Int64
	errorCount         atomic.Int64
	latencies          []time.Duration
	expectedMaxSuccess int
}

func NewMetricsCollector(expectedMaxSuccess int) *MetricsCollector {
	return &MetricsCollector{
		latencies:          make([]time.Duration, 0, 1000),
		expectedMaxSuccess: expectedMaxSuccess,
	}
}

// Record records a single request result
func (mc *MetricsCollector) Record(statusCode int, err error, latency time.Duration) {
	if err != nil {
		mc.errorCount.Add(1)
		return
	}

	mc.mu.Lock()
	mc.latencies = append(mc.latencies, latency)
	mc.mu.Unlock()

	switch statusCode {
	case 200:
		mc.successCount.Add(1)
	case 429:
		mc.rateLimitCount.Add(1)
	default:
		mc.errorCount.Add(1)
	}
}

func (mc *MetricsCollector) GenerateResult(duration time.Duration, scenario string) TestResult {
	successCount := int(mc.successCount.Load())
	rateLimitCount := int(mc.rateLimitCount.Load())
	errorCount := int(mc.errorCount.Load())
	totalRequests := successCount + rateLimitCount + errorCount

	// Calculate race condition detection
	overLimitGrants := 0
	raceDetected := false
	if successCount > mc.expectedMaxSuccess {
		overLimitGrants = successCount - mc.expectedMaxSuccess
		raceDetected = true
	}

	// Calculate throughput
	throughput := float64(totalRequests) / duration.Seconds()

	latency := mc.calculateLatencyMetrics()

	return TestResult{
		Scenario:           scenario,
		TotalRequests:      totalRequests,
		SuccessCount:       successCount,
		RateLimitCount:     rateLimitCount,
		ErrorCount:         errorCount,
		Duration:           duration,
		Throughput:         throughput,
		Latency:            latency,
		ExpectedMaxSuccess: mc.expectedMaxSuccess,
		OverLimitGrants:    overLimitGrants,
		RaceDetected:       raceDetected,
	}
}

// calculateLatencyMetrics computes percentile latencies
func (mc *MetricsCollector) calculateLatencyMetrics() LatencyMetrics {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if len(mc.latencies) == 0 {
		return LatencyMetrics{}
	}

	sort.Slice(mc.latencies, func(i, j int) bool {
		return mc.latencies[i] < mc.latencies[j]
	})

	n := len(mc.latencies)

	p50 := mc.latencies[n*50/100]
	p95 := mc.latencies[n*95/100]
	p99 := mc.latencies[n*99/100]
	min := mc.latencies[0]
	max := mc.latencies[n-1]

	var sum time.Duration
	for _, lat := range mc.latencies {
		sum += lat
	}
	avg := sum / time.Duration(n)

	return LatencyMetrics{
		P50: p50,
		P95: p95,
		P99: p99,
		Max: max,
		Min: min,
		Avg: avg,
	}
}
