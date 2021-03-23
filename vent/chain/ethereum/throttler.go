package ethereum

import (
	"math/big"
	"time"

	"github.com/hyperledger/burrow/logging"
)

type Throttler struct {
	// Request timestamps as unix nanos (avoid space overhead of time.Time)
	requests                 []int64
	maxRequestsPerNanosecond *big.Float
	// Window over which to accumulate request times
	window time.Duration
	logger *logging.Logger
}

func NewThrottler(maxRequests int, timeBase time.Duration, window time.Duration, logger *logging.Logger) *Throttler {
	maxRequestsPerNanosecond := new(big.Float).SetInt64(int64(maxRequests))
	maxRequestsPerNanosecond.Quo(maxRequestsPerNanosecond, new(big.Float).SetInt64(int64(timeBase)))
	return &Throttler{
		maxRequestsPerNanosecond: maxRequestsPerNanosecond,
		window:                   window,
		logger:                   logger,
	}
}

func (t *Throttler) Throttle() {
	time.Sleep(t.calculateWait())
}

func (t *Throttler) calculateWait() time.Duration {
	requests := len(t.requests)
	if requests < 2 {
		return 0
	}
	delta := t.requests[requests-1] - t.requests[0]
	deltaNanoseconds := new(big.Float).SetInt64(delta)

	allowedRequestsInDelta := new(big.Float).Mul(deltaNanoseconds, t.maxRequestsPerNanosecond)

	overage := allowedRequestsInDelta.Sub(new(big.Float).SetInt64(int64(requests)), allowedRequestsInDelta)
	if overage.Sign() > 0 {
		// Wait just long enough to eat our overage at max request rate
		nanos, _ := new(big.Float).Quo(overage, t.maxRequestsPerNanosecond).Int64()
		wait := time.Duration(nanos)
		t.logger.InfoMsg("throttling connection",
			"num_requests", requests,
			"over_period", time.Duration(delta).String(),
			"requests_overage", overage.String(),
			"wait", wait.String(),
		)
		return wait
	}
	return 0
}

func (t *Throttler) addNow() {
	t.add(time.Now())
}

func (t *Throttler) add(now time.Time) {
	cutoff := now.Add(-t.window)
	// Remove expired requests
	for len(t.requests) > 0 && t.requests[0] < cutoff.UnixNano() {
		t.requests = t.requests[1:]
	}
	t.requests = append(t.requests, now.UnixNano())
}
