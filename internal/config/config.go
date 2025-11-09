package config

import (
	"fmt"

	"github.com/redis/go-redis/v9"
)

type ServerConfig struct {
	Port       string
	InstanceId string
}

type LimiterConfig struct {
	Strategy       string
	StrategyConfig map[string]interface{}
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	PoolSize int
}

type Redis struct {
	Config RedisConfig
	Client *redis.Client
}

type Config struct {
	Server  ServerConfig
	Limiter LimiterConfig
	Redis   Redis
}

func Load() (*Config, error) {
	strategy := getEnvString("LIMITER_STRATEGY", "tokenbucket-lua")

	// this is as per selected strategy
	strategyConfig := map[string]interface{}{
		"capacity":   getEnvInt64("LIMITER_CAPACITY", 20),
		"refillRate": getEnvInt64("LIMITER_REFILL_RATE", 5),
	}

	if strategy == "tokenbucket-transaction" {
		strategyConfig["maxRetries"] = getEnvInt("MAX_RETRIES", 10)
	}

	cfg := &Config{
		Server: ServerConfig{
			Port:       getEnvString("PORT", "1783"),
			InstanceId: getInstanceId(),
		},
		Limiter: LimiterConfig{
			Strategy:       strategy,
			StrategyConfig: strategyConfig,
		},
		Redis: Redis{
			Config: RedisConfig{
				Addr:     getEnvString("REDIS_ADDR", "localhost:6379"),
				Password: getEnvString("REDIS_PASSWORD", ""),
				DB:       getEnvInt("REDIS_DB", 0),
				PoolSize: getEnvInt("REDIS_POOL_SIZE", 20),
			},
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cfg *Config) Validate() error {
	if cfg.Server.Port == "" {
		return fmt.Errorf("port cannot be empty")
	}
	if cfg.Server.InstanceId == "" {
		return fmt.Errorf("instanceId cannot be empty")
	}

	if cfg.Redis.Config.Addr == "" {
		return fmt.Errorf("redis address cannot be empty")
	}
	if cfg.Redis.Config.DB < 0 {
		return fmt.Errorf("redis database cannot be negative")
	}
	if cfg.Redis.Config.PoolSize <= 0 {
		return fmt.Errorf("redis pool size must be positive")
	}

	return nil
}
