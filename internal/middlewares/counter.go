package middlewares

import (
	"log"
	"net/http"
	"time"

	"github.com/brutally-Honest/distributed-rate-limiter/internal/redis"
)

func Counter() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx:=r.Context()
			next.ServeHTTP(w, r)
			log.Printf("%s %s (%v)", r.Method, r.URL.Path, duration)
		})
	}
}
