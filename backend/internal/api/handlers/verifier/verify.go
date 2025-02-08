package verifier

import (
	"app-pinger/backend/internal/api/utilapi"
	"crypto/subtle"
	"errors"
	"net/http"
	"time"
)

func (v *Verifier) Verify(ctx *utilapi.APIContext) {
	receivedKey := ctx.GetFromHeader("X-API-Key")
	auth := false
	for _, key := range v.Auth.apiKeys {
		if subtle.ConstantTimeCompare([]byte(receivedKey), []byte(key)) == 1 {
			auth = true
			break
		}
	}

	if !auth {
		ctx.Error("unauthorized request", errors.New("invalid API-Key"))
		ctx.WriteFailure(http.StatusUnauthorized, "invalid request")
		return
	}

	v.RateLimiter.mu.Lock()
	defer v.RateLimiter.mu.Unlock()
	ip := ctx.GetFromHeader("X-Real-IP")
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

	time.AfterFunc(v.RateLimiter.rateTime, func() {
		v.RateLimiter.mu.Lock()
		defer v.RateLimiter.mu.Unlock()
		v.RateLimiter.rateLimiter[ip] = 1
		v.RateLimiter.lastAccess[ip] = time.Now()
	})
}
