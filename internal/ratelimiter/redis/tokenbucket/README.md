# Redis Token Bucket Implementations

This directory contains multiple approaches for implementing token bucket rate limiting in Redis, each with different performance and consistency trade-offs.

## Implementation 1: Redis Hashes

**File**: `hash.go`

### Approach

Uses Redis hash data structures (`HSET`/`HMGET`) for storing token bucket state. Reads current tokens and last refill timestamp, calculates token refill, updates both values in separate operations. Prefers hashes over simple strings for better organization and extensibility.

### Benefits

- Simple implementation with minimal code complexity
- Single network round-trip for reading bucket state
- Clean namespace organization with structured data
- Easy to extend with additional metadata fields

### Trade-offs

- **Race conditions**: Multiple instances can read stale state simultaneously, allowing over-limit requests under high concurrency
- **Inconsistent behavior**: Token over-allocation increases with load and distributed instances
- **Non-atomic operations**: Read and write happen separately, creating race condition windows
- **Eventual consistency**: Behavior degrades under concurrent access

## Implementation 2: Redis Transactions

**File**: `transaction.go`

### Approach

Uses Redis `WATCH`/`MULTI`/`EXEC` commands for optimistic locking to ensure atomic read-modify-write operations through transactions with automatic retry logic on transaction conflicts.

### Benefits

- Eliminates race conditions through transactional atomicity
- Maintains hash-based storage advantages
- No changes to data structure or API

### Trade-offs

- Higher latency due to potential transaction retries
- More complex error handling for failed transactions
- Increased Redis command overhead

## Implementation 3: Lua Script

**File**: `lua.go`

### Approach

Server-side Lua script execute atomically in Redis with single round-trip for complete rate limit check and update cycle. Script handles all logic: token calculation, refill, consumption.

### Benefits

- Zero race conditions - atomic execution guaranteed
- Best performance with minimal network overhead
- Consistent behavior across all distributed instances

### Trade-offs

- Higher complexity in script development and testing
- Increased Redis server CPU usage
- More challenging debugging compared to client-side logic

## Implementation Comparison

| Aspect              | Hash-based | Transactions      | Lua Script    |
| ------------------- | ---------- | ----------------- | ------------- |
| Race Conditions     | ⚠️ Present | ✅ Eliminated     | ✅ Eliminated |
| Performance         | Highest    | Good              | Best          |
| Complexity          | Low        | Medium            | High          |
| Redis Load          | Low        | Medium            | Medium        |
| Consistency         | Eventual   | Strong            | Strong        |
| Network Round-trips | 2          | 3+ (with retries) | 1             |

## Current Status

All three implementations are **fully functional**:

- ✅ **Hash-based** (`hash.go`) - Basic implementation with race conditions
- ✅ **Transaction-based** (`transaction.go`) - Atomic with retry logic
- ✅ **Lua-based** (`lua.go`) - Recommended for most use cases
