package redis

import (
	"encoding/json"
	"fmt"

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
		fmt.Println("Strategy selected: ", strategy)
		fmt.Println("Strategy config: ", strategyConfig)
	}()
	switch strategy {

	case "tokenbucket":
		var cfg tokenbucket.TBConfig
		configBytes, _ := json.Marshal(strategyConfig)
		if err := json.Unmarshal(configBytes, &cfg); err != nil {
			return nil, fmt.Errorf("invalid token bucket config: %w", err)
		}
		if err := cfg.Validate(); err != nil {
			return nil, err
		}
		return tokenbucket.NewTBHash(client, cfg, instanceId), nil

	default:
		return nil, fmt.Errorf("unknown strategy: %s", strategy)
	}
}
