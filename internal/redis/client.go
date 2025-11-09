package redis

import (
	"context"
	"fmt"
	"log"

	"github.com/brutally-Honest/distributed-rate-limiter/internal/config"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
	Config config.RedisConfig
}

func New(redisConfig config.RedisConfig) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Addr,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
		PoolSize: redisConfig.PoolSize,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}
	log.Println("Redis client created successfully")
	return &RedisClient{
		Client: rdb,
		Config: redisConfig,
	}, nil
}

func (r *RedisClient) GetClient() *redis.Client {
	return r.Client
}

func (r *RedisClient) Close() error {
	return r.Client.Close()
}
