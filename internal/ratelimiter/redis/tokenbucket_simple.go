package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenBucketNaive implements RateLimiter using separate Redis commands
// WARNING: This has race conditions under concurrent load
type TokenBucketSimple struct {
	client *redis.Client
	config TokenBucketConfig
}

func NewTokenBucketSimple(client *redis.Client, cfg TokenBucketConfig) *TokenBucketSimple {
	return &TokenBucketSimple{
		client: client,
		config: cfg,
	}
}

// CheckLimit checks if the request should be allowed
// RACE CONDITION: Multiple requests can read the same value simultaneously
// remaining represents tokens left in the bucket
func (tb *TokenBucketSimple) CheckLimit(ctx context.Context, key string) (bool, int, error) {
	fullKey := fmt.Sprintf("ratelimit:tb:%s", key)
	now := time.Now().Unix()

	// RACE: Read current state
	bucket, err := tb.client.HMGet(ctx, fullKey, "tokens", "last_refill").Result()
	if err != nil && err != redis.Nil {
		return false, 0, fmt.Errorf("redis hmget failed: %w", err)
	}

	var currentTokens float64
	var lastRefill int64

	// Initialize if doesn't exist
	if bucket[0] == nil {
		currentTokens = float64(tb.config.Capacity)
		lastRefill = now
	} else {
		currentTokens, err = strconv.ParseFloat(bucket[0].(string), 64)
		if err != nil {
			return false, 0, fmt.Errorf("failed to parse tokens: %w", err)
		}
		lastRefill, err = strconv.ParseInt(bucket[1].(string), 10, 64)
		if err != nil {
			return false, 0, fmt.Errorf("failed to parse last_refill: %w", err)
		}
	}

	// Calculate refill
	elapsed := now - lastRefill
	tokensToAdd := float64(elapsed) * float64(tb.config.RefillRate)
	currentTokens = min(float64(tb.config.Capacity), currentTokens+tokensToAdd)

	allowed := currentTokens >= 1.0
	if allowed {
		currentTokens -= 1.0
	}

	// RACE: Write new state (another request may have modified between GET and SET)
	err = tb.client.HMSet(ctx, fullKey,
		"tokens", fmt.Sprintf("%.2f", currentTokens),
		"last_refill", now,
	).Err()

	if err != nil {
		return false, 0, fmt.Errorf("redis hmset failed: %w", err)
	}

	// Set expiration
	tb.client.Expire(ctx, fullKey, time.Hour)

	return allowed, int(currentTokens), nil
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
