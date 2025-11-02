# Redis Token Bucket Implementations

Different approaches for implementing token bucket rate limiting in Redis

## Implementation 1: Hash-based

**File**: `tokenbucket_simple.go`

### Design Choice: Hashes vs Simple Strings

**Why Redis Hashes (`HSET`/`HMGET`) over `GET`/`SET`:**

**Benefits:**
- **Atomicity within a key**: Single Redis key stores both `tokens` and `last_refill` fields atomically
- **Cleaner namespace organization**: One bucket key instead of multiple separate keys
- **Extensibility**: Easy to add additional bucket metadata without key proliferation  
- **Performance**: Single network round-trip for reading complete bucket state

**Trade-offs considered:**
- Slightly higher memory usage per bucket
- More complex parsing compared to simple float strings
- But negligible overhead for rate limiting workloads

### Observations
**Problems:**
- **Race conditions:** Multiple distributed instances can simultaneously read the same token bucket state, leading to inconsistent rate limiting behavior which increases with higher concurrency and load.
- **Token over-allocation:** Allows more requests than the configured rate limit permits.

**Vulnerability Points:**
- **Read Operation**: `HMGET` retrieves bucket state (non-atomic with subsequent operations)
- **Processing Gap**: Time between reading state and writing updates allows other instances to read stale data
- **Concurrent Refill Calculations**: Multiple instances may add tokens based on the same `last_refill` timestamp


## Implementation 2: Redis Transactions (Planned)

<!-- **File**: `tokenbucket_transaction.go`

### Approach
- Uses Redis `WATCH`/`MULTI`/`EXEC` commands for optimistic locking
- Ensures atomic read-modify-write operations through transactions
- Automatic retry logic on transaction conflicts

### Expected Benefits
- Eliminates race conditions through transactional atomicity
- Maintains hash-based storage advantages
- No changes to data structure or API

### Trade-offs
- Higher latency due to potential transaction retries
- More complex error handling for failed transactions
- Increased Redis command overhead -->

## Implementation 3: Lua Scripts (Planned)

<!-- **File**: `tokenbucket_lua.go`

### Approach
- Server-side Lua scripts execute atomically in Redis
- Single round-trip for complete rate limit check and update cycle
- Script handles all logic: token calculation, refill, consumption

### Expected Benefits
- Zero race conditions - atomic execution guaranteed
- Best performance with minimal network overhead
- Consistent behavior across all distributed instances

### Trade-offs
- Higher complexity in script development and testing
- Increased Redis server CPU usage
- More challenging debugging compared to client-side logic

## Implementation Comparison

| Aspect | Simple Hash | Transactions | Lua Scripts |
|--------|-------------|--------------|-------------|
| Race Conditions | Present | Eliminated | Eliminated |
| Performance | Highest | Good | Good |
| Complexity | Low | Medium | High |
| Redis Load | Low | Medium | Medium |
| Consistency | Eventual | Strong | Strong |
| Network Round-trips | 2 | 3+ (with retries) | 1 |

## Usage Guidelines

**Choose Implementation 1 (Simple)** when:
- Low to medium concurrency
- Performance is critical
- Some over-limit requests are acceptable
- Simplicity outweighs perfect accuracy

**Choose Implementation 2 (Transactions)** when:
- High concurrency with moderate consistency requirements
- Need stronger guarantees than simple approach
- Can tolerate occasional transaction retry latency

**Choose Implementation 3 (Lua Scripts)** when:
- Maximum consistency and accuracy required
- High concurrency with zero tolerance for race conditions
- Performance and atomicity are both critical requirements

## Future Considerations

- Benchmark all implementations under various load patterns
- Consider Redis Cluster deployment implications
- Evaluate memory usage patterns across implementations
- Monitor for Redis command overhead differences -->