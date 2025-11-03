package tokenbucket

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type TBHash struct {
	client     *redis.Client
	config     TBConfig
	instanceId string
}

func NewTBHash(client *redis.Client, cfg TBConfig, instanceId string) *TBHash {
	return &TBHash{
		client:     client,
		config:     cfg,
		instanceId: instanceId,
	}
}

// RACE CONDITION: Multiple requests can read the same value simultaneously
func (tb *TBHash) CheckLimit(ctx context.Context, key string) (bool, int, error) {
	fullKey := fmt.Sprintf("ratelimit:tb:%s", key)

	// calculation with seconds
	nowSeconds := time.Now().Unix()

	// precise logging with nanoseconds
	nowNano := time.Now().UnixNano()

	// RACE: Read current state
	bucket, err := tb.client.HMGet(ctx, fullKey, "tokens", "last_refill").Result()
	log.Printf("[%s][%d ns] Checking Redis bucket: %v\n", tb.instanceId, nowNano, bucket)
	if err != nil && err != redis.Nil {
		return false, 0, fmt.Errorf("redis hmget failed: %w", err)
	}
	var currentTokens float64
	var lastRefillSeconds int64

	if bucket[0] == nil {
		currentTokens = float64(tb.config.Capacity)
		lastRefillSeconds = nowSeconds
	} else {
		currentTokens, err = strconv.ParseFloat(bucket[0].(string), 64)
		if err != nil {
			return false, 0, fmt.Errorf("failed to parse tokens: %w", err)
		}
		lastRefillSeconds, err = strconv.ParseInt(bucket[1].(string), 10, 64)
		if err != nil {
			return false, 0, fmt.Errorf("failed to parse last_refill: %w", err)
		}
	}

	elapsedSeconds := nowSeconds - lastRefillSeconds
	tokensToAdd := float64(elapsedSeconds) * float64(tb.config.RefillRate)
	currentTokens = min(float64(tb.config.Capacity), currentTokens+tokensToAdd)
	log.Printf("[%s][%d ns] Calculated: elapsed=%ds, tokensToAdd=%.2f, currentTokens=%.2f\n",
		tb.instanceId,
		time.Now().UnixNano(), elapsedSeconds, tokensToAdd, currentTokens)
	allowed := currentTokens >= 1.0
	if allowed {
		currentTokens -= 1.0
	}
	// RACE: Write new state
	err = tb.client.HSet(ctx, fullKey,
		"tokens", fmt.Sprintf("%.2f", currentTokens),
		"last_refill", nowSeconds,
	).Err()

	log.Printf("[%s][%d ns] Setting Redis bucket: %.2f (allowed=%v)\n",
		tb.instanceId,
		time.Now().UnixNano(), currentTokens, allowed)

	if err != nil {
		return false, 0, fmt.Errorf("redis hset failed: %w", err)
	}
	tb.client.Expire(ctx, fullKey, time.Hour)
	return allowed, int(currentTokens), nil
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
