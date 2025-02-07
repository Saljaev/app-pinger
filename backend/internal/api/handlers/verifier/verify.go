package verifier

import (
	"app-pinger/backend/internal/api/utilapi"
	"net/http"
	"time"
)

func (v *Verifier) Verify(ctx *utilapi.APIContext) {
	//TODO: make auth
	v.RateLimiter.mu.Lock()
	defer v.RateLimiter.mu.Unlock()
	ip := ctx.GetFromQuery("X-Real-IP")
	if count, ok := v.RateLimiter.rateLimiter[ip]; ok {
		if count > v.RateLimiter.rateLimit {
			ctx.Debug("to many request from", "IP", ip)
			ctx.WriteFailure(http.StatusTooManyRequests, "server is busy")
			return
		}
		v.RateLimiter.rateLimiter[ip]++
	} else {
		v.RateLimiter.rateLimiter[ip] = 1
		v.RateLimiter.lastAccess[ip] = time.Now()
	}

	if time.Now().Sub(v.RateLimiter.lastAccess[ip]) > v.RateLimiter.rateTime {
		v.RateLimiter.rateLimiter[ip] = 1
		v.RateLimiter.lastAccess[ip] = time.Now()
	}
}
