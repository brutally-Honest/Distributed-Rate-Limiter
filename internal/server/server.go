package server

import (
	"fmt"
	"net/http"

	"github.com/brutally-Honest/distributed-rate-limiter/internal/config"
	"github.com/brutally-Honest/distributed-rate-limiter/internal/middlewares"
	"github.com/brutally-Honest/distributed-rate-limiter/internal/ratelimiter"
	"github.com/brutally-Honest/distributed-rate-limiter/internal/redis"
)

type Server struct {
	httpServer  *http.Server
	config      *config.Config
	redisClient *redis.RedisClient
}

func New(cfg *config.Config) (*Server, error) {

	redisClient, err := redis.New(
		cfg.Redis.Addr,
		cfg.Redis.Password,
		cfg.Redis.DB,
		cfg.Redis.PoolSize,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis client: %v", err)
	}

	limiter, err := ratelimiter.NewRateLimiter(cfg.Limiter.Strategy, cfg.Limiter.StrategyConfig, redisClient.GetClient(), cfg.Server.InstanceId)
	if err != nil {
		return nil, fmt.Errorf("failed to create rate limiter: %v", err)
	}
	router := SetUpRoutes(cfg)
	handlersWithMiddleware := middlewares.Chain(
		middlewares.Logger(),
		middlewares.NewRateLimiterMiddleware(limiter),
	)(router)

	s := &Server{
		config:      cfg,
		redisClient: redisClient,
		httpServer: &http.Server{
			Addr:    ":" + cfg.Server.Port,
			Handler: handlersWithMiddleware,
		},
	}

	return s, nil
}

func (s *Server) Start() error {
	fmt.Println("Server running on port", s.config.Server.Port)
	return s.httpServer.ListenAndServe()
}
