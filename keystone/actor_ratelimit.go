package keystone

import (
	"context"
	"github.com/keystonedb/sdk-go/proto"
)

type RateLimit struct {
	key          string // The key to rate limit on
	hardLimit    int32  // The number of requests allowed in the rate limit period
	limitMinutes int32  // The number of minutes to check the rate limit over
	actor        *Actor
	readDistinct bool
	historical   bool
}

type RateLimitResult struct {
	currentCount int32
	hitLimit     bool
	percent      float64
}

func (r RateLimitResult) CurrentCount() int32 { return r.currentCount }
func (r RateLimitResult) HitLimit() bool      { return r.hitLimit }
func (r RateLimitResult) Percent() float64    { return r.percent }

func (r *RateLimit) Trigger(ctx context.Context, transactionId string) (RateLimitResult, error) {
	resp := RateLimitResult{}
	conn := r.actor.Connection()
	res, err := conn.RateLimit(ctx, &proto.RateLimitRequest{
		Authorization:   r.actor.Authorization(),
		Key:             r.key,
		HardLimit:       r.hardLimit,
		RateMinutes:     r.limitMinutes,
		TransactionId:   transactionId,
		ReadDistinct:    r.readDistinct,
		StoreHistorical: r.historical,
	})

	if err == nil && res != nil {
		resp.currentCount = res.GetCurrentCount()
		resp.hitLimit = res.GetOverLimit()
		resp.percent = float64(resp.currentCount) / float64(r.hardLimit)
	}

	return resp, err
}

func (a *Actor) newRateLimit(key string, hardLimit, limitMinutes int32, historical, distinct bool) *RateLimit {
	return &RateLimit{
		actor:        a,
		key:          key,
		hardLimit:    hardLimit,
		limitMinutes: limitMinutes,
		readDistinct: distinct,
		historical:   historical,
	}
}

func (a *Actor) NewRateLimit(key string, hardLimit, limitMinutes int32) *RateLimit {
	return a.newRateLimit(key, hardLimit, limitMinutes, false, true)
}

func (a *Actor) NewTrackedRateLimit(key string, hardLimit, limitMinutes int32) *RateLimit {
	return a.newRateLimit(key, hardLimit, limitMinutes, true, true)
}
