package load

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"

	"testing"
	"time"
)

func testSingleClientBurst(t *testing.T, strategy string) TestResult {
	t.Helper()

	const (
		numRequests = 200
		clientIP    = "192.168.1.100"
		concurrency = 50
	)

	collector := NewMetricsCollector(testCapacity)
	start := time.Now()

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrency)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		semaphore <- struct{}{}

		go func() {
			defer wg.Done()
			defer func() { <-semaphore }()

			reqStart := time.Now()
			statusCode, err := makeRequest(clientIP)
			latency := time.Since(reqStart)

			collector.Record(statusCode, err, latency)
		}()
	}

	wg.Wait()
	duration := time.Since(start)

	return collector.GenerateResult(duration, "Single Client Burst")
}

func testMultipleClientsConcurrent(t *testing.T, strategy string) TestResult {
	t.Helper()

	const (
		numClients        = 5
		requestsPerClient = 30
		concurrency       = 25
	)

	totalRequests := numClients * requestsPerClient
	expectedSuccess := numClients * testCapacity
	if expectedSuccess > totalRequests {
		expectedSuccess = totalRequests
	}

	collector := NewMetricsCollector(expectedSuccess)
	start := time.Now()

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrency)

	for clientID := 0; clientID < numClients; clientID++ {
		clientIP := fmt.Sprintf("192.168.1.%d", 100+clientID)

		for req := 0; req < requestsPerClient; req++ {
			wg.Add(1)
			semaphore <- struct{}{}

			go func(ip string) {
				defer wg.Done()
				defer func() { <-semaphore }()

				reqStart := time.Now()
				statusCode, err := makeRequest(ip)
				latency := time.Since(reqStart)

				collector.Record(statusCode, err, latency)
			}(clientIP)
		}
	}

	wg.Wait()
	duration := time.Since(start)

	return collector.GenerateResult(duration, "Multiple Clients Concurrent")
}

// Tests: Token refill behavior and recovery after exhaustion
// Expected: Initial burst exhausts bucket, refill allows more requests
func testBurstThenSteady(t *testing.T, strategy string) TestResult {
	t.Helper()

	const (
		burstSize      = 120 // Exhaust bucket
		steadyRate     = 5   // Requests per second
		steadyDuration = 3 * time.Second
		clientIP       = "192.168.1.200"
		concurrency    = 30
	)

	// Calculate expected success
	// Burst: min(burstSize, capacity)
	// Steady: refillRate * steadyDuration
	expectedBurst := testCapacity
	expectedSteady := testRefillRate * int(steadyDuration.Seconds())
	expectedSuccess := expectedBurst + expectedSteady

	collector := NewMetricsCollector(expectedSuccess)
	start := time.Now()

	// Phase 1: Burst
	t.Logf("Phase 1: Sending burst of %d requests", burstSize)
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrency)

	for i := 0; i < burstSize; i++ {
		wg.Add(1)
		semaphore <- struct{}{}

		go func() {
			defer wg.Done()
			defer func() { <-semaphore }()

			reqStart := time.Now()
			statusCode, err := makeRequest(clientIP)
			latency := time.Since(reqStart)

			collector.Record(statusCode, err, latency)
		}()
	}

	wg.Wait()
	t.Logf("Phase 1 complete. Waiting for token refill...")

	// Phase 2: Wait for refill
	time.Sleep(1 * time.Second)

	// Phase 3: Steady traffic
	t.Logf("Phase 2: Sending steady traffic at %d req/s for %v", steadyRate, steadyDuration)
	ticker := time.NewTicker(time.Second / time.Duration(steadyRate))
	defer ticker.Stop()

	steadyCtx, cancel := context.WithTimeout(context.Background(), steadyDuration)
	defer cancel()

	for {
		select {
		case <-steadyCtx.Done():
			goto done
		case <-ticker.C:
			wg.Add(1)
			go func() {
				defer wg.Done()

				reqStart := time.Now()
				statusCode, err := makeRequest(clientIP)
				latency := time.Since(reqStart)

				collector.Record(statusCode, err, latency)
			}()
		}
	}

done:
	wg.Wait()
	duration := time.Since(start)

	return collector.GenerateResult(duration, "Burst Then Steady")
}

func makeRequest(clientIP string) (int, error) {
	req, err := http.NewRequest("GET", baseURL+"/api", nil)
	if err != nil {
		return 0, err
	}

	// Simulate client IP via X-Forwarded-For
	req.Header.Set("X-Forwarded-For", clientIP)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	io.Copy(io.Discard, resp.Body)

	return resp.StatusCode, nil
}
