package redis

import (
	"encoding/json"
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

	var cfg tokenbucket.TBConfig

	switch strategy {
	case "tokenbucket-hash":
		if err := decodeConfig(strategyConfig, &cfg); err != nil {
			return nil, err
		}
		return tokenbucket.NewTBHash(client, cfg, instanceId), nil

	case "tokenbucket-transaction":
		if err := decodeConfig(strategyConfig, &cfg); err != nil {
			return nil, err
		}
		maxRetries := 3
		if retries, ok := strategyConfig["maxRetries"].(float64); ok {
			maxRetries = int(retries)
		}
		return tokenbucket.NewTBTransaction(client, cfg, maxRetries), nil

	case "tokenbucket-lua":
		if err := decodeConfig(strategyConfig, &cfg); err != nil {
			return nil, err
		}
		return tokenbucket.NewTBLua(client, cfg), nil
	default:
		return nil, fmt.Errorf("unknown strategy: %s", strategy)
	}
}

func decodeConfig(strategyConfig map[string]interface{}, cfg *tokenbucket.TBConfig) error {
	configBytes, err := json.Marshal(strategyConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal strategy config: %w", err)
	}
	if err := json.Unmarshal(configBytes, cfg); err != nil {
		return fmt.Errorf("invalid token bucket config: %w", err)
	}
	return cfg.Validate()
}
