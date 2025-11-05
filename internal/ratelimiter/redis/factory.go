package redis

import (
	"fmt"
	"log"

	"github.com/brutally-Honest/distributed-rate-limiter/internal/ratelimiter"
	"github.com/brutally-Honest/distributed-rate-limiter/internal/ratelimiter/redis/tokenbucket"
	"github.com/redis/go-redis/v9"
)

func NewRateLimiter(
	strategy string,
	strategyConfig map[string]interface{},
	client *redis.Client,
	instanceId string) (ratelimiter.RateLimiter, error) {

	defer func() {
		log.Printf("Rate limiter initialized | strategy=%s | config=%+v", strategy, strategyConfig)
	}()

	switch strategy {
	case "tokenbucket-hash", "tokenbucket-transaction", "tokenbucket-lua":
		var cfg tokenbucket.TBConfig
		if err := decodeTBConfig(strategyConfig, &cfg); err != nil {
			return nil, err
		}

		switch strategy {
		case "tokenbucket-hash":
			return tokenbucket.NewTBHash(client, cfg, instanceId), nil

		case "tokenbucket-transaction":
			maxRetries := 3
			if retries, ok := strategyConfig["maxRetries"].(int); ok {
				maxRetries = retries
			}
			return tokenbucket.NewTBTransaction(client, cfg, maxRetries), nil

		case "tokenbucket-lua":
			return tokenbucket.NewTBLua(client, cfg), nil

		default:
			return nil, fmt.Errorf("unknown strategy: %s", strategy)
		}
	}
	return nil, fmt.Errorf("unknown strategy: %s", strategy)
}

func decodeTBConfig(strategyConfig map[string]interface{}, cfg *tokenbucket.TBConfig) error {
	capacity, ok := strategyConfig["capacity"].(int64)
	if !ok {
		return fmt.Errorf("capacity must be int64, got %T", strategyConfig["capacity"])
	}
	cfg.Capacity = capacity

	refillRate, ok := strategyConfig["refillRate"].(int64)
	if !ok {
		return fmt.Errorf("refillRate must be int64, got %T", strategyConfig["refillRate"])
	}
	cfg.RefillRate = refillRate

	return cfg.Validate()
}
