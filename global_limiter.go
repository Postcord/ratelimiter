package ratelimiter

import (
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type GlobalLimiter struct {
	sync.RWMutex
	resetAfter float64
	logger     zerolog.Logger
}

type Reservation struct {
	wait time.Duration
}

func (r *Reservation) Delay() time.Duration {
	return r.wait
}

func (g *GlobalLimiter) Reserve() *Reservation {
	g.RLock()
	defer g.RUnlock()
	return &Reservation{
		wait: time.Duration(g.resetAfter * float64(time.Second)),
	}
}

func (g *GlobalLimiter) UpdateResetAfter(resetAfter float64) {
	g.Lock()
	defer g.Unlock()
	if resetAfter > 0 {
		g.resetAfter = resetAfter
	}
	go func() {
		g.Lock()
		defer g.Unlock()
		log.Warn().Dur("reset_after", time.Duration(g.resetAfter*float64(time.Second))).Msg("Globally ratelimited")
		time.Sleep(time.Duration(g.resetAfter * float64(time.Second)))
		g.resetAfter = 0
	}()
}

func NewGlobalLimiter(logger zerolog.Logger) *GlobalLimiter {
	return &GlobalLimiter{
		resetAfter: 0,
		logger:     logger,
	}
}
