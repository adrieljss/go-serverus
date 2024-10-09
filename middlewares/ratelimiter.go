package middlewares

import (
	"time"

	"github.com/adrieljansen/go-serverus/result"
	"github.com/adrieljansen/go-serverus/utils"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

// rate limiter based on IP
type RateLimitInstance struct {
	limiter *rate.Limiter
}

type IPRateLimiter struct {
	IPRateLimiterMap    *utils.TtlMap[string, RateLimitInstance]
	RateLimitBucketSize int
	RefillFrequency     rate.Limit
}

var IPRateLimiterCache *IPRateLimiter

// bind to variable
func StartIPRateLimiterService(refillFrequency rate.Limit, bucketSize int) *IPRateLimiter {
	IPRateLimiterCache = &IPRateLimiter{
		IPRateLimiterMap:    utils.NewLastAccessTtlMap[string, RateLimitInstance](time.Hour * 2),
		RefillFrequency:     refillFrequency,
		RateLimitBucketSize: bucketSize,
	}
	logrus.Warn("initated ttl map for ip rate limiter")
	return IPRateLimiterCache
}

func (i *IPRateLimiter) AddIP(ip string) *rate.Limiter {
	rl := rate.NewLimiter(IPRateLimiterCache.RefillFrequency, IPRateLimiterCache.RateLimitBucketSize)
	i.IPRateLimiterMap.Store(ip, RateLimitInstance{
		limiter: rl,
	}, int64(time.Minute)*40)
	return rl
}

func (i *IPRateLimiter) Allow(ip string) bool {
	inst, ok := i.IPRateLimiterMap.Get(ip)
	if ok {
		return inst.limiter.Allow()
	} else {
		// not exist, then add
		i.AddIP(ip)
		return true
	}
}

// use this as a middleware if the route is ratelimited
//
// uses IP-Based ratelimiting
func RateLimitRequired() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !IPRateLimiterCache.Allow(ctx.ClientIP()) {
			result.Err(429, nil, "TOO_MANY_REQUESTS", "too many requests in a short amount of time").SendJSON(ctx)
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
