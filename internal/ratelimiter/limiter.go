package ratelimiter

import "context"

type RateLimiter interface {
	CheckLimit(ctx context.Context, ip string) (allowed bool, remaining int, error error)
}
