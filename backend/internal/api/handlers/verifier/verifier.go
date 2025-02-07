package verifier

import (
	"sync"
	"time"
)

type Verifier struct {
	RateLimiter RateLimiter
	Auth        Auth
}

type RateLimiter struct {
	rateLimiter map[string]int
	lastAccess  map[string]time.Time
	mu          sync.Mutex
	rateLimit   int
	rateTime    time.Duration
}

type Auth struct {
	apiKeys []string
}

func NewVerifier(k []string, rL int, rT time.Duration) *Verifier {
	return &Verifier{
		RateLimiter: RateLimiter{
			rateLimiter: make(map[string]int),
			lastAccess:  make(map[string]time.Time),
			mu:          sync.Mutex{},
			rateLimit:   rL,
			rateTime:    rT,
		},
		Auth: Auth{
			apiKeys: k,
		},
	}
}
