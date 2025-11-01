package ratelimiter

import (
	"encoding/json"
	"fmt"

	rl "github.com/brutally-Honest/distributed-rate-limiter/internal/ratelimiter/redis"
	"github.com/redis/go-redis/v9"
)

func NewRateLimiter(strategy string, strategyConfig map[string]interface{}, client *redis.Client) (RateLimiter, error) {
	defer func() {
		fmt.Println("Strategy selected: ", strategy)
		fmt.Println("Strategy config: ", strategyConfig)
	}()
	switch strategy {

	case "tokenbucket":
		var cfg rl.TokenBucketConfig
		configBytes, _ := json.Marshal(strategyConfig)
		if err := json.Unmarshal(configBytes, &cfg); err != nil {
			return nil, fmt.Errorf("invalid token bucket config: %w", err)
		}
		if err := cfg.Validate(); err != nil {
			return nil, err
		}
		return rl.NewTokenBucketSimple(client, cfg), nil

	default:
		return nil, fmt.Errorf("unknown strategy: %s", strategy)
	}
}
