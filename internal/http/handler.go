package handlers

import (
	"time"

	"github.com/brutally-Honest/distributed-rate-limiter/internal/config"
)

type Handlers struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Handlers {
	return &Handlers{
		cfg: cfg,
	}
}

type Resp struct {
	Msg        string    `json:"msg"`
	Time       time.Time `json:"time"`
	InstanceId string    `json:"instance_id"`
}
