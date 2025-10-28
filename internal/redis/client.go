// package redis

// import (
//     "os"
//     "sync"
//     "github.com/redis/go-redis/v9"
// )

// var (
//     client *redis.Client
//     once   sync.Once
// )

// func GetClient() *redis.Client {
//     once.Do(func() {
//         addr := os.Getenv("REDIS_ADDR")
//         if addr == "" {
//             addr = "redis:6379"
//         }
//         client = redis.NewClient(&redis.Options{
//             Addr:     addr,
//             Password: "",
//             DB:       0,
//             PoolSize: 20,
//         })
//     })
//     return client
// }

package redis

import (
    "context"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

type RedisClient struct {
    client *redis.Client
}

func New(addr string) *RedisClient {
    rdb := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: "",
        DB:       0,
        PoolSize: 20,
    })
    return &RedisClient{client: rdb}
}

// Increment a key with expiration
func (r *RedisClient) Increment(ctx context.Context, key string) (int64, error) {
    count, err := r.client.Incr(ctx, key).Result()
    if err != nil {
        return 0, err
    }

    return count, nil
}
