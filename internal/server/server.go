package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/brutally-Honest/distributed-rate-limiter/internal/config"
	"github.com/brutally-Honest/distributed-rate-limiter/internal/middlewares"
	redis_ratelimiter "github.com/brutally-Honest/distributed-rate-limiter/internal/ratelimiter/redis"
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
		return nil, fmt.Errorf("failed to create Redis client: %w", err)
	}

	limiter, err := redis_ratelimiter.NewRateLimiter(
		cfg.Limiter.Strategy,
		cfg.Limiter.StrategyConfig,
		redisClient.GetClient(),
		cfg.Server.InstanceId,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create rate limiter: %w", err)
	}

	limiterMiddleware := middlewares.NewRateLimiterMiddleware(limiter)
	router := SetUpRoutes(cfg, limiterMiddleware)

	handlersWithMiddleware := middlewares.Chain(
		middlewares.Logger(),
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
	log.Println("Server running on port", s.config.Server.Port)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Preparing to shutdown server...")

	// TODO: Make this configurable
	shutdownCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
		return err
	}

	if err := s.redisClient.Close(); err != nil {
		log.Printf("Redis client shutdown error: %v", err)
		return err
	}

	log.Printf("Server %s shutdown complete", s.config.Server.InstanceId)
	return nil
}
