package handlers

import (
	"time"

	"github.com/brutally-Honest/distributed-rate-limiter/internal/config"
)

type Handlers struct {
	cfg       *config.Config
	startTime time.Time
}

func New(cfg *config.Config, startTime time.Time) *Handlers {
	return &Handlers{
		cfg:       cfg,
		startTime: startTime,
	}
}
