package tokenbucket

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const tokenBucketLuaScript = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local requested = tonumber(ARGV[3])
local now = tonumber(ARGV[4])
local bucket = redis.call('HMGET', key, 'tokens', 'last_refill')
local tokens = tonumber(bucket[1])
local last_refill = tonumber(bucket[2])
-- Initialize if doesn't exist
if tokens == nil then
    tokens = capacity
    last_refill = now
end
-- Calculate refill
local elapsed = now - last_refill
local tokens_to_add = elapsed * refill_rate
tokens = math.min(capacity, tokens + tokens_to_add)
-- Try to consume
if tokens >= requested then
    tokens = tokens - requested
    redis.call('HMSET', key, 'tokens', tokens, 'last_refill', now)
    redis.call('EXPIRE', key, 3600)
    return {1, math.floor(tokens)}
else
    redis.call('HMSET', key, 'tokens', tokens, 'last_refill', now)
    redis.call('EXPIRE', key, 3600)
    return {0, math.floor(tokens)}
end
`

type TBLua struct {
	client *redis.Client
	config TBConfig
	script *redis.Script
}

func NewTBLua(client *redis.Client, cfg TBConfig) *TBLua {
	return &TBLua{
		client: client,
		config: cfg,
		script: redis.NewScript(tokenBucketLuaScript),
	}
}

func (tb *TBLua) CheckLimit(ctx context.Context, key string) (bool, int, error) {
	fullKey := fmt.Sprintf("ratelimit:tb:%s", key)
	now := float64(time.Now().Unix())
	result, err := tb.script.Run(
		ctx,
		tb.client,
		[]string{fullKey},
		tb.config.Capacity,
		tb.config.RefillRate,
		1,
		now,
	).Int64Slice()
	if err != nil {
		return false, 0, fmt.Errorf("redis lua script failed: %w", err)
	}
	if len(result) != 2 {
		return false, 0, fmt.Errorf("unexpected lua script result length: %d", len(result))
	}
	allowed := result[0] == 1
	remaining := int(result[1])
	return allowed, remaining, nil
}
