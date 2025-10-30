package config

import (
	"fmt"
)

type ServerConfig struct {
	Port       string
	InstanceId string
}

type LimiterConfig struct {
	Capacity   int64
	RefillRate int64
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	PoolSize int
}

type Config struct {
	Server  ServerConfig
	Limiter LimiterConfig
	Redis   RedisConfig
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:       getEnvString("PORT", "1783"),
			InstanceId: getInstanceId(),
		},
		Limiter: LimiterConfig{
			Capacity:   getEnvInt64("LIMITER_CAPACITY", 20),
			RefillRate: getEnvInt64("LIMITER_REFILL_RATE", 5),
		},
		Redis: RedisConfig{
			Addr:     getEnvString("REDIS_ADDR", "localhost:6379"),
			Password: getEnvString("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
			PoolSize: getEnvInt("REDIS_POOL_SIZE", 20),
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

	if cfg.Limiter.Capacity <= 0 {
		return fmt.Errorf("rate limiter capacity must be positive")
	}
	if cfg.Limiter.RefillRate <= 0 {
		return fmt.Errorf("rate limiter refill rate must be positive")
	}

	if cfg.Redis.Addr == "" {
		return fmt.Errorf("redis address cannot be empty")
	}
	if cfg.Redis.Password == "" {
		return fmt.Errorf("redis password cannot be empty")
	}
	if cfg.Redis.DB < 0 {
		return fmt.Errorf("redis database cannot be negative")
	}
	if cfg.Redis.PoolSize <= 0 {
		return fmt.Errorf("redis pool size must be positive")
	}

	return nil
}
