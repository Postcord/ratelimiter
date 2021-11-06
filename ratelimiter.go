package ratelimiter

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Ratelimiter struct {
	collector *Collector
	logger    zerolog.Logger
}

func NewRatelimiter(logger ...zerolog.Logger) *Ratelimiter {
	if len(logger) == 0 {
		logger = []zerolog.Logger{log.Logger}
	}
	return &Ratelimiter{
		collector: NewCollector(logger[0]),
		logger:    logger[0].With().Str("component", "ratelimiter").Logger(),
	}
}

func (r *Ratelimiter) Limit(req *http.Request) error {
	res := r.collector.global.Reserve()
	time.Sleep(res.Delay())
	log.Debug().Msg("Got global reservation")
	bucket, id := r.collector.GetBucket(req.URL.Path)
	if bucket != nil {
		res := bucket.Reserve()
		if !res.OK() {
			return NewErrUnavailable("not OK")
		}
		time.Sleep(res.Delay())
		log.Debug().Str("bucket", id).Msg("Got bucket reservation")
	}
	return nil
}

func (r *Ratelimiter) Update(resp *http.Response) {
	r.collector.UpdateFromResponse(resp)
	log.Debug().Msg("Updated from response")
}
