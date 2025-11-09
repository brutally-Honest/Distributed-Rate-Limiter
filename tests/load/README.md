# Load Testing Results

## Running Tests

### Prerequisites


```bash
docker-compose up --build --scale 'go=x'
```
### Run Tests


```bash
# Test all strategies (run these one at a time, restarting server with different LIMITER_STRATEGY)
cd tests/load


# Test Hash strategy
go test -v -run TestHashStrategy


# Test Transaction strategy  
go test -v -run TestTransactionStrategy


# Test Lua strategy
go test -v -run TestLuaStrategy
```
### Notes  
- Tune number of instances and strategy configuraion as needed 
- x is the desired number of instances

## Metrics Collected

For each test run:
- Throughput: Requests per second
- Latency: p50, p95, p99, max
- Success Rate: 200 responses vs 429 rate limits
- Race Detection: Over-limit grants (actual success > expected)

## Test Setup

### Test Environment:
- Setup: 4 service instances behind Nginx, shared Redis
- Configuration: Capacity=50 tokens, Refill Rate=10 tokens/second

### Test Scenarios:
1. Single Client Burst: 200 concurrent requests from one IP (tests race conditions)
2. Multiple Clients Concurrent: 5 clients × 30 requests each (tests fairness)
3. Burst Then Steady: Initial burst of 120, then 5 req/s for 3s (tests refill behavior)


## Observations


### Hash Strategy Findings

```
go test -v -run TestHashStrategy
=== RUN   TestHashStrategy
    load_test.go:27: Testing Hash-based Token Bucket Strategy
=== RUN   TestHashStrategy/Single_Client_Burst
redis: 2025/11/07 06:17:42 redis.go:478: auto mode fallback: maintnotifications disabled due to handshake error: ERR unknown subcommand 'maint_notifications'. Try CLIENT HELP.
    load_test.go:64: 
        ========================================
        Strategy: tokenbucket-hash
        Scenario: Single Client Burst
        ========================================
        Total Requests: 200
        Successful (200): 200
        Rate Limited (429): 0
        Errors: 0
        Duration: 49.670541ms
        Throughput: 4026.53 req/sec
        ----------------------------------------
        Latency p50: 7.112916ms
        Latency p95: 27.957167ms
        Latency p99: 29.64675ms
        Latency max: 29.685042ms
        ----------------------------------------
        Race Condition Check:
        Expected Max Success: 50
        Actual Success: 200
        Over-limit Grants: 150
        Race Detected: true
        ========================================
=== RUN   TestHashStrategy/Multiple_Clients_Concurrent
    load_test.go:64: 
        ========================================
        Strategy: tokenbucket-hash
        Scenario: Multiple Clients Concurrent
        ========================================
        Total Requests: 150
        Successful (200): 150
        Rate Limited (429): 0
        Errors: 0
        Duration: 28.010625ms
        Throughput: 5355.11 req/sec
        ----------------------------------------
        Latency p50: 3.596125ms
        Latency p95: 11.036542ms
        Latency p99: 17.602ms
        Latency max: 19.886959ms
        ----------------------------------------
        Race Condition Check:
        Expected Max Success: 150
        Actual Success: 150
        Over-limit Grants: 0
        Race Detected: false
        ========================================
=== RUN   TestHashStrategy/Burst_Then_Steady
    load_test.go:58: Phase 1: Sending burst of 120 requests
    load_test.go:58: Phase 1 complete. Waiting for token refill...
    load_test.go:58: Phase 2: Sending steady traffic at 5 req/s for 3s
    load_test.go:64: 
        ========================================
        Strategy: tokenbucket-hash
        Scenario: Burst Then Steady
        ========================================
        Total Requests: 135
        Successful (200): 135
        Rate Limited (429): 0
        Errors: 0
        Duration: 4.019269542s
        Throughput: 33.59 req/sec
        ----------------------------------------
        Latency p50: 2.832167ms
        Latency p95: 5.020125ms
        Latency p99: 6.334833ms
        Latency max: 7.237125ms
        ----------------------------------------
        Race Condition Check:
        Expected Max Success: 80
        Actual Success: 135
        Over-limit Grants: 55
        Race Detected: true
        ========================================
=== RUN   TestHashStrategy/Summary
    load_test.go:70: 
        ########################################
        # STRATEGY SUMMARY: tokenbucket-hash
        ########################################
    load_test.go:70: Single Client Burst            | Throughput: 4026.53 req/s | p95: 27.957167ms | Races: true
    load_test.go:70: Multiple Clients Concurrent    | Throughput: 5355.11 req/s | p95: 11.036542ms | Races: false
    load_test.go:70: Burst Then Steady              | Throughput:  33.59 req/s | p95: 5.020125ms | Races: true
    load_test.go:70: ########################################
        
--- PASS: TestHashStrategy (4.41s)
    --- PASS: TestHashStrategy/Single_Client_Burst (0.16s)
    --- PASS: TestHashStrategy/Multiple_Clients_Concurrent (0.13s)
    --- PASS: TestHashStrategy/Burst_Then_Steady (4.12s)
    --- PASS: TestHashStrategy/Summary (0.00s)
PASS
ok  	github.com/brutally-Honest/distributed-rate-limiter/tests/load	4.804s
```
- Race Conditions Confirmed: CATASTROPHIC failure under load
  - Single Client Burst: 200/50 succeeded (300% over-limit)
  - Burst Then Steady: 135/80 succeeded (169% over-limit)
  - Total over-grants: 205 requests that should have been blocked
- Root Cause: Separate HGET and HSET operations create race window where multiple instances read stale state simultaneously
- Performance: 4027 req/s throughput, p95 latency 28ms
- Verdict: UNSAFE for production - will leak rate limits under any concurrent load


### Transaction Strategy Findings
```
go test -v -run TestTransactionStrategy
=== RUN   TestTransactionStrategy
    load_test.go:31: Testing Transaction-based Token Bucket Strategy
=== RUN   TestTransactionStrategy/Single_Client_Burst
redis: 2025/11/07 06:29:36 redis.go:478: auto mode fallback: maintnotifications disabled due to handshake error: ERR unknown subcommand 'maint_notifications'. Try CLIENT HELP.
    load_test.go:64: 
        ========================================
        Strategy: tokenbucket-transaction
        Scenario: Single Client Burst
        ========================================
        Total Requests: 200
        Successful (200): 50
        Rate Limited (429): 150
        Errors: 0
        Duration: 87.312166ms
        Throughput: 2290.63 req/sec
        ----------------------------------------
        Latency p50: 13.511209ms
        Latency p95: 56.345208ms
        Latency p99: 59.306667ms
        Latency max: 59.587958ms
        ----------------------------------------
        Race Condition Check:
        Expected Max Success: 50
        Actual Success: 50
        Over-limit Grants: 0
        Race Detected: false
        ========================================
=== RUN   TestTransactionStrategy/Multiple_Clients_Concurrent
    load_test.go:64: 
        ========================================
        Strategy: tokenbucket-transaction
        Scenario: Multiple Clients Concurrent
        ========================================
        Total Requests: 150
        Successful (200): 142
        Rate Limited (429): 8
        Errors: 0
        Duration: 45.045ms
        Throughput: 3330.00 req/sec
        ----------------------------------------
        Latency p50: 5.897833ms
        Latency p95: 16.463875ms
        Latency p99: 18.394708ms
        Latency max: 19.149084ms
        ----------------------------------------
        Race Condition Check:
        Expected Max Success: 150
        Actual Success: 142
        Over-limit Grants: 0
        Race Detected: false
        ========================================
=== RUN   TestTransactionStrategy/Burst_Then_Steady
    load_test.go:58: Phase 1: Sending burst of 120 requests
    load_test.go:58: Phase 1 complete. Waiting for token refill...
    load_test.go:58: Phase 2: Sending steady traffic at 5 req/s for 3s
    load_test.go:64: 
        ========================================
        Strategy: tokenbucket-transaction
        Scenario: Burst Then Steady
        ========================================
        Total Requests: 135
        Successful (200): 65
        Rate Limited (429): 70
        Errors: 0
        Duration: 4.050792292s
        Throughput: 33.33 req/sec
        ----------------------------------------
        Latency p50: 8.384875ms
        Latency p95: 19.087875ms
        Latency p99: 21.118292ms
        Latency max: 22.123459ms
        ----------------------------------------
        Race Condition Check:
        Expected Max Success: 80
        Actual Success: 65
        Over-limit Grants: 0
        Race Detected: false
        ========================================
=== RUN   TestTransactionStrategy/Summary
    load_test.go:70: 
        ########################################
        # STRATEGY SUMMARY: tokenbucket-transaction
        ########################################
    load_test.go:70: Single Client Burst            | Throughput: 2290.63 req/s | p95: 56.345208ms | Races: false
    load_test.go:70: Multiple Clients Concurrent    | Throughput: 3330.00 req/s | p95: 16.463875ms | Races: false
    load_test.go:70: Burst Then Steady              | Throughput:  33.33 req/s | p95: 19.087875ms | Races: false
    load_test.go:70: ########################################
        
--- PASS: TestTransactionStrategy (4.50s)
    --- PASS: TestTransactionStrategy/Single_Client_Burst (0.20s)
    --- PASS: TestTransactionStrategy/Multiple_Clients_Concurrent (0.15s)
    --- PASS: TestTransactionStrategy/Burst_Then_Steady (4.15s)
    --- PASS: TestTransactionStrategy/Summary (0.00s)
PASS
ok  	github.com/brutally-Honest/distributed-rate-limiter/tests/load	4.968s
```

- Race Conditions: Zero over-limit grants across all scenarios
- Atomicity: WATCH/MULTI/EXEC prevents concurrent modifications
- Performance Impact: 
  - Throughput: 2291 req/s (43% slower than Hash)
  - p95 latency: 56ms (2x Hash latency)
  - Why: Transaction retries on conflicts add overhead
- Behavior: Slightly conservative (65/80 in Burst Then Steady, likely due to retry timeouts)
- Verdict: Correct but expensive

### Lua Strategy Findings
```
go test -v -run TestLuaStrategy
=== RUN   TestLuaStrategy
    load_test.go:35: Testing Lua Script-based Token Bucket Strategy
=== RUN   TestLuaStrategy/Single_Client_Burst
redis: 2025/11/07 06:31:30 redis.go:478: auto mode fallback: maintnotifications disabled due to handshake error: ERR unknown subcommand 'maint_notifications'. Try CLIENT HELP.
    load_test.go:64: 
        ========================================
        Strategy: tokenbucket-lua
        Scenario: Single Client Burst
        ========================================
        Total Requests: 200
        Successful (200): 50
        Rate Limited (429): 150
        Errors: 0
        Duration: 42.881291ms
        Throughput: 4664.04 req/sec
        ----------------------------------------
        Latency p50: 7.095333ms
        Latency p95: 22.181334ms
        Latency p99: 22.834959ms
        Latency max: 23.296875ms
        ----------------------------------------
        Race Condition Check:
        Expected Max Success: 50
        Actual Success: 50
        Over-limit Grants: 0
        Race Detected: false
        ========================================
=== RUN   TestLuaStrategy/Multiple_Clients_Concurrent
    load_test.go:64: 
        ========================================
        Strategy: tokenbucket-lua
        Scenario: Multiple Clients Concurrent
        ========================================
        Total Requests: 150
        Successful (200): 150
        Rate Limited (429): 0
        Errors: 0
        Duration: 17.758916ms
        Throughput: 8446.46 req/sec
        ----------------------------------------
        Latency p50: 2.572667ms
        Latency p95: 4.939417ms
        Latency p99: 5.370083ms
        Latency max: 5.41ms
        ----------------------------------------
        Race Condition Check:
        Expected Max Success: 150
        Actual Success: 150
        Over-limit Grants: 0
        Race Detected: false
        ========================================
=== RUN   TestLuaStrategy/Burst_Then_Steady
    load_test.go:58: Phase 1: Sending burst of 120 requests
    load_test.go:58: Phase 1 complete. Waiting for token refill...
    load_test.go:58: Phase 2: Sending steady traffic at 5 req/s for 3s
    load_test.go:64: 
        ========================================
        Strategy: tokenbucket-lua
        Scenario: Burst Then Steady
        ========================================
        Total Requests: 135
        Successful (200): 65
        Rate Limited (429): 70
        Errors: 0
        Duration: 4.025371292s
        Throughput: 33.54 req/sec
        ----------------------------------------
        Latency p50: 2.999792ms
        Latency p95: 11.641834ms
        Latency p99: 12.627708ms
        Latency max: 12.646792ms
        ----------------------------------------
        Race Condition Check:
        Expected Max Success: 80
        Actual Success: 65
        Over-limit Grants: 0
        Race Detected: false
        ========================================
=== RUN   TestLuaStrategy/Summary
    load_test.go:70: 
        ########################################
        # STRATEGY SUMMARY: tokenbucket-lua
        ########################################
    load_test.go:70: Single Client Burst            | Throughput: 4664.04 req/s | p95: 22.181334ms | Races: false
    load_test.go:70: Multiple Clients Concurrent    | Throughput: 8446.46 req/s | p95: 4.939417ms | Races: false
    load_test.go:70: Burst Then Steady              | Throughput:  33.54 req/s | p95: 11.641834ms | Races: false
    load_test.go:70: ########################################
        
--- PASS: TestLuaStrategy (4.40s)
    --- PASS: TestLuaStrategy/Single_Client_Burst (0.15s)
    --- PASS: TestLuaStrategy/Multiple_Clients_Concurrent (0.12s)
    --- PASS: TestLuaStrategy/Burst_Then_Steady (4.13s)
    --- PASS: TestLuaStrategy/Summary (0.00s)
PASS
ok  	github.com/brutally-Honest/distributed-rate-limiter/tests/load	4.575s
```

- Race Conditions: Zero over-limit grants across all scenarios
- Atomicity: Single server-side script execution, no race windows
- Performance: Best of all strategies
  - Single Client Burst: 4664 req/s (16% FASTER than Hash!)
  - Multi-Client: 8446 req/s (58% faster than Hash, 154% faster than Transaction)
  - p95 latency: 22ms (21% better than Hash, 61% better than Transaction)
- Why so fast: Single Redis round-trip, no client-side calculation overhead, no retries needed
- Verdict: Combines consistency with highest performance


## Performance Comparison Table


| Scenario | Metric | Hash | Transaction | Lua | Winner |
|----------|--------|------|-------------|-----|--------|
| Single Client Burst | Throughput | 4027 req/s | 2291 req/s | 4664 req/s | Lua |
| | p95 Latency | 28ms | 56ms | 22ms | Lua |
| | Success Rate | 200/50 | 50/50 | 50/50 | Transaction/Lua |
| | Over-limit Grants | 150 | 0 | 0 | Transaction/Lua |
| Multi-Client | Throughput | 5355 req/s | 3330 req/s | 8446 req/s | Lua |
| | p95 Latency | 11ms | 16ms | 5ms | Lua |
| | Success Rate | 150/150 | 142/150 | 150/150 | Hash/Lua |
| Burst Then Steady | Throughput | 34 req/s | 33 req/s | 34 req/s | Tie |
| | p95 Latency | 5ms | 19ms | 12ms | Hash |
| | Success Rate | 135/80 | 65/80 | 65/80 | Transaction/Lua |
| | Over-limit Grants | 55 | 0 | 0 | Transaction/Lua |


Summary: Lua strategy achieves perfect consistency (0 race conditions) while outperforming even the broken Hash strategy by 16% in high-concurrency scenarios.

## Key Takeaways

1. Hash strategy is unsafe - 205 over-limit requests granted (300%+ error rate)
2. Lua strategy outperforms even the broken Hash strategy - atomic AND 16% faster
3. Transaction strategy works but at 2x latency cost - acceptable fallback if Lua unavailable
4. Real distributed race conditions measured: 4 instances competing for shared state exposed Hash failures immediately


## Test Limitations

- Tests run against local Redis (production latency will be higher)
- Network latency within Docker environment (lower than real distributed systems)
- Simplified load patterns (real traffic is more complex)

#### For true production validation, run these tests:
- Against production-like Redis cluster with network latency
- With service instances on separate machines (not Docker containers on same host)
- Over extended duration (minutes to hours, not seconds)
- With realistic traffic patterns from your use case