package ratelimiter

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/time/rate"
)

type Collector struct {
	sync.RWMutex
	buckets  map[string]*rate.Limiter
	routeMap map[string]string
	global   *GlobalLimiter
	logger   zerolog.Logger
}

func NewCollector(logger zerolog.Logger) *Collector {
	return &Collector{
		buckets:  make(map[string]*rate.Limiter),
		routeMap: make(map[string]string),
		global:   NewGlobalLimiter(logger),
		logger:   logger,
	}
}

func (c *Collector) BucketExists(name string) bool {
	c.RLock()
	defer c.RUnlock()
	_, ok := c.buckets[name]
	return ok
}

func (c *Collector) mappingExists(name string) bool {
	c.RLock()
	defer c.RUnlock()
	_, ok := c.routeMap[name]
	return ok
}

func (c *Collector) addBucket(name string, r time.Duration, count int) {
	c.Lock()
	defer c.Unlock()
	c.logger.Debug().Msgf("Adding bucket %s with %d requests per %f seconds", name, count, r)
	c.buckets[name] = rate.NewLimiter(rate.Every(r), count)
	// Reserve a ticket since we JUST made a request
	c.buckets[name].Reserve()
}

func (c *Collector) addMapping(routeKey, bucket string) {
	c.Lock()
	defer c.Unlock()
	c.logger.Debug().Msgf("Adding mapping %s to bucket %s", routeKey, bucket)
	c.routeMap[routeKey] = bucket
}

func (c *Collector) GetBucket(path string) (*rate.Limiter, string) {
	c.RLock()
	defer c.RUnlock()
	routeKey := parseRouteKey(path)
	bucketID := c.routeMap[routeKey]
	b, ok := c.buckets[bucketID]
	if !ok {
		return nil, ""
	}
	return b, bucketID
}

func (c *Collector) UpdateFromResponse(r *http.Response) {
	h := r.Header
	bucket := h.Get("X-RateLimit-Bucket")
	limitHeader := h.Get("X-RateLimit-Limit")
	resetAfter := h.Get("X-RateLimit-Reset-After")
	global := h.Get("X-RateLimit-Global")

	count, err := strconv.ParseInt(limitHeader, 10, 64)
	if err != nil {
		return
	}

	reset, err := strconv.ParseFloat(resetAfter, 64)
	if err != nil {
		return
	}

	if global != "" {
		c.global.UpdateResetAfter(reset)
	} else {

		if !c.BucketExists(bucket) {
			// Upon first request, limit and resetAfter should be actual values
			c.addBucket(bucket, time.Duration(reset)*time.Second, int(count))
		}

		routeKey := parseRouteKey(r.Request.URL.Path)

		if !c.mappingExists(routeKey) {
			c.addMapping(routeKey, bucket)
		}
	}
}

var snowRe = regexp.MustCompile(`\d{17,19}`)

func parseRouteKey(path string) string {
	splitPath := strings.Split(path, "/")
	includeNext := true
	routeKeyParts := []string{}
	for _, c := range splitPath[2:] {
		isSnowflake := snowRe.MatchString(c)
		if isSnowflake && includeNext {
			routeKeyParts = append(routeKeyParts, c)
			includeNext = false
		} else if !isSnowflake {
			routeKeyParts = append(routeKeyParts, c)
			if c == "channel" || c == "guild" || c == "webhooks" {
				includeNext = true
			}
		}
	}
	return strings.Join(routeKeyParts, ":")
}
