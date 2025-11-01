package middlewares

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

type RateLimiterConfig struct {
}

func RateLimiter() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := extractIP(r)
			fmt.Println("ip extracted", ip)
			next.ServeHTTP(w, r)
		})
	}
}

func extractIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	fmt.Println("remote addr", r.RemoteAddr)
	fmt.Println("xff", xff)
	if xff != "" {
		ips := strings.Split(xff, ",")
		fmt.Println("ips", ips)
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	// fallback to remote addr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil || ip == "" {
		ip = r.RemoteAddr
	}
	return ip
}
