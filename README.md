# Distributed Rate Limiter 

A distributed rate limiter built with Go and Redis, implementing token bucket algorithm for microservices and API gateways. Features clean architecture, dependency injection, and extensible design for adding new rate limiting strategies.

[![Go Version](https://img.shields.io/badge/go-1.24.5-blue.svg)](https://golang.org)

## Overview

This project demonstrates Go development practices for building distributed systems. Currently implements a hash-based token bucket algorithm with Redis for distributed rate limiting across multiple service instances.

**Key Features:**
- **Distributed**: Scales across multiple instances with Redis coordination
- **Configurable**: Environment-based configuration with validation
- **Observable**: Structured logging with instance identification
- **Performant**: Hash-based Redis operations for efficient rate limiting
- **Extensible**: Factory pattern enables easy addition of new rate limiting strategies

## Architecture

### Clean Architecture Principles

```
cmd/server/           # Application entry point
├── main.go          # Bootstrap and dependency injection

internal/
├── config/          # Configuration management
├── server/          # HTTP server setup and routing
├── middlewares/     # HTTP middleware chain
├── ratelimiter/     # Rate limiting abstractions
│   ├── limiter.go   # RateLimiter interface
│   └── redis/       # Redis-based implementations
│       ├── factory.go      # Rate limiter factory
│       └── tokenbucket/    # Token bucket implementations
│           ├── config.go   # Token bucket configuration
│           ├── hash.go     # Hash-based token bucket
│           └── README.md   # Implementation details
├── redis/           # Redis client wrapper
└── http/            # HTTP handlers
```

### Design Patterns

**Dependency Injection:**
- Constructor injection for all dependencies
- Interface-based design enabling testability
- Clean separation of concerns

**Factory Pattern:**
- Strategy-based rate limiter creation
- Extensible algorithm implementations
- Configuration-driven instantiation

**Middleware Chain:**
- Composable HTTP middleware
- Clean request/response flow
- Cross-cutting concerns (logging, rate limiting)

**Adapter Pattern:**
- Redis client abstraction
- External dependency isolation
- Test-friendly interfaces

### Rate Limiting Flow

```mermaid
graph LR
    A[HTTP Request] --> B[Rate Limit Middleware]
    B --> C{Extract IP/Key}
    C --> D[Check Rate Limit]
    D --> E{Allowed?}
    E -->|Yes| F[Proceed to Handler]
    E -->|No| G[Return 429]
    D --> H[Redis Operations]
```

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.24.5+ (for development)

### Run with Docker Compose

```bash
# Clone and navigate to project
git clone <repository>
cd distributed-rate-limiter

# Start all services
docker-compose up --build

# Test rate limiting
curl http://localhost/api
hey -n 50 -c 10 http://localhost/api
```

### Local Development

```bash
# Install dependencies
go mod download

# Run Redis
docker run -d -p 6379:6379 redis:7-alpine

# Set environment variables
export PORT=1783
export LIMITER_CAPACITY=20
export LIMITER_REFILL_RATE=5
export REDIS_ADDR=localhost:6379

# Run the server
go run cmd/server/main.go
```

## ⚙️ Configuration

Environment-based configuration with validation:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `1783` | HTTP server port |
| `LIMITER_STRATEGY` | `tokenbucket` | Rate limiting algorithm |
| `LIMITER_CAPACITY` | `20` | Token bucket capacity |
| `LIMITER_REFILL_RATE` | `5` | Tokens added per second |
| `REDIS_ADDR` | `localhost:6379` | Redis connection address |
| `REDIS_PASSWORD` | `""` | Redis password |
| `REDIS_DB` | `0` | Redis database number |
| `REDIS_POOL_SIZE` | `20` | Redis connection pool size |

## 🔧 Go Best Practices Demonstrated

### Idiomatic Go

**Project Structure:**
- `internal/` package for private APIs
- Clear package naming and organization
- Single responsibility principle

**Error Handling:**
- Error wrapping with context (`fmt.Errorf("failed: %w", err)`)
- Structured error propagation
- Graceful degradation

**Concurrency:**
- Context-aware operations
- Proper resource cleanup
- Thread-safe Redis operations

### Performance Optimizations

**Redis Efficiency:**
- Connection pooling with configurable size
- Hash-based storage for atomic field operations
- Minimal network round-trips

**Memory Management:**
- Efficient string parsing
- Minimal allocations in hot paths
- Proper resource lifecycle management

### Dependency Management

**Minimal External Dependencies:**
- Only Redis client (`github.com/redis/go-redis/v9`)
- No additional frameworks or libraries
- Standard library heavy for core functionality

**Clean Interfaces:**
```go
type RateLimiter interface {
    CheckLimit(ctx context.Context, key string) (allowed bool, remaining int, err error)
}
```

## 📊 API Endpoints

### GET /api
Rate-limited endpoint returning JSON response.

**Response:**
```json
{
  "msg": "Successfully Hit",
  "time": "2024-01-01T12:00:00Z",
  "instanceId": "instance-123"
}
```

**Rate Limit Headers:**
- `X-RateLimit-Remaining`: Tokens remaining in bucket

### GET /health
Health check endpoint.

## 🧪 Testing & Quality

### Current Test Coverage
- [ ] Unit tests for rate limiting logic
- [ ] Integration tests with Redis
- [ ] Load testing scenarios
- [ ] Race condition verification tests

### Planned Enhancements
- [ ] Custom error types with structured error codes
- [ ] Log aggregation to database
- [ ] Metrics collection and monitoring
- [ ] Configuration validation improvements

## 🔍 Implementation Details

### Token Bucket Algorithm

The token bucket algorithm provides smooth rate limiting:

- **Capacity**: Maximum burst allowance
- **Refill Rate**: Tokens added per second
- **Distributed State**: Redis maintains consistent state across instances

### Race Condition Analysis

The current hash-based implementation has known race conditions where multiple instances can read stale token bucket state simultaneously. See [`internal/ratelimiter/redis/tokenbucket/README.md`](internal/ratelimiter/redis/tokenbucket/README.md) for detailed analysis of concurrency issues and planned atomic implementations.

### IP Extraction Strategy

Intelligent client IP detection:
1. `X-Forwarded-For` header (first IP in chain)
2. `RemoteAddr` fallback
3. Proper IPv4/IPv6 handling

## 🔄 Development Workflow

```bash
# Run tests
go test ./...

# Run with race detection
go run -race cmd/server/main.go

# Format code
go fmt ./...

# Lint code
go vet ./...

# Build for production
go build -o bin/server cmd/server/main.go
```
## 🚢 Deployment

### Docker Production Build

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server ./cmd/server

FROM alpine:latest
COPY --from=builder /app/server .
CMD ["./server"]
```

## 📈 Performance Benchmarks

**Target Metrics:**
- < 1ms average response time
- 99th percentile < 5ms
- Support 10k+ RPS per instance

**Current Implementation:**
- Hash-based token bucket using Redis HSET/HMGET operations
- Atomic field updates within Redis keys
- Single network round-trip per rate limit check

## 📚 Learnings & Patterns

This project demonstrates:

- **Clean Architecture** with proper separation of concerns
- **Dependency Injection** for testable code
- **Factory Pattern** for extensible algorithm implementations
- **Middleware Pattern** for composable HTTP handling
- **Interface-based Design** for maintainable code

## 🔗 Related Documentation

- [Token Bucket Implementation Details](internal/ratelimiter/redis/tokenbucket/README.md)
- [Redis Client Documentation](internal/redis/client.go)

---

**Built with ❤️ and Go best practices**

