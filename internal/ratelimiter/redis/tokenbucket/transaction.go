package tokenbucket

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type TBTransaction struct {
	client     *redis.Client
	config     TBConfig
	maxRetries int
}

func NewTBTransaction(client *redis.Client, cfg TBConfig, maxRetries int) *TBTransaction {
	return &TBTransaction{
		client:     client,
		config:     cfg,
		maxRetries: maxRetries,
	}
}

// Retries on transaction conflicts (WATCH failures)
func (tb *TBTransaction) CheckLimit(ctx context.Context, key string) (bool, int, error) {
	fullKey := fmt.Sprintf("ratelimit:tb:%s", key)
	now := time.Now().Unix()
	var allowed bool
	var remaining float64

	// Retries on transaction conflicts (WATCH failures)
	for attempt := 0; attempt < tb.maxRetries; attempt++ {
		err := tb.client.Watch(ctx, func(tx *redis.Tx) error {

			bucket, err := tx.HMGet(ctx, fullKey, "tokens", "last_refill").Result()
			if err != nil && err != redis.Nil {
				return fmt.Errorf("hmget failed: %w", err)
			}
			var currentTokens float64
			var lastRefill int64

			if bucket[0] == nil {
				currentTokens = float64(tb.config.Capacity)
				lastRefill = now
			} else {
				currentTokens, err = strconv.ParseFloat(bucket[0].(string), 64)
				if err != nil {
					return fmt.Errorf("parse tokens failed: %w", err)
				}
				lastRefill, err = strconv.ParseInt(bucket[1].(string), 10, 64)
				if err != nil {
					return fmt.Errorf("parse last_refill failed: %w", err)
				}
			}

			elapsed := now - lastRefill
			tokensToAdd := float64(elapsed) * float64(tb.config.RefillRate)
			currentTokens = min(float64(tb.config.Capacity), currentTokens+tokensToAdd)

			allowed = currentTokens >= 1.0
			if allowed {
				currentTokens -= 1.0
			}
			remaining = currentTokens
			// Execute transaction atomically
			_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				pipe.HSet(ctx, fullKey,
					"tokens", fmt.Sprintf("%.2f", currentTokens),
					"last_refill", now,
				)
				pipe.Expire(ctx, fullKey, time.Hour)
				return nil
			})
			return err
		}, fullKey)
		if err == nil {
			return allowed, int(remaining), nil
		}
		if err == redis.TxFailedErr {
			continue
		}
		return false, 0, fmt.Errorf("redis transaction failed: %w", err)
	}
	fmt.Printf("Max retries exceeded for key: %s\n", fullKey)
	return false, 0, nil
}
