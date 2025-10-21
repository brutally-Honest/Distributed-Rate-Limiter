package config

import "fmt"

type ServerConfig struct {
	Port       string
	InstanceId string
}

type LimiterConfig struct {
	Capacity   int64
	RefillRate int64
}

type Config struct {
	Server  ServerConfig
	Limiter LimiterConfig
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:       "1783",
			InstanceId: "local-1",
		},
		Limiter: LimiterConfig{
			Capacity:   20,
			RefillRate: 5,
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
		return fmt.Errorf("rate limiter capacity cannot be empty")
	}
	if cfg.Limiter.RefillRate <= 0 {
		return fmt.Errorf("rate limiter refill rate cannot be empty")
	}
	return nil
}
