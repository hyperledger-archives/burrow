package ethereum

import (
	"testing"
	"time"

	"github.com/hyperledger/burrow/logging/logconfig"

	"github.com/stretchr/testify/assert"
)

var logger = logconfig.Sink().Terminal().LoggingConfig().WithTrace().MustLogger()

const delta = float64(time.Millisecond)

func TestThrottler_Overage(t *testing.T) {
	throttler := NewThrottler(100, time.Second, time.Minute, logger)

	now := doRequests(throttler, 100, time.Now(), time.Second)
	assert.InDelta(t, time.Duration(0), throttler.calculateWait(), delta)

	doRequests(throttler, 200, now, time.Second)
	assert.InDelta(t, time.Second, throttler.calculateWait(), delta)
}

func TestThrottler_Expiry(t *testing.T) {
	throttler := NewThrottler(100, time.Second, 2*time.Second, logger)

	now := doRequests(throttler, 200, time.Now(), time.Second)
	assert.InDelta(t, time.Second, throttler.calculateWait(), delta)

	now = doRequests(throttler, 100, now, time.Second)
	assert.InDelta(t, time.Second, throttler.calculateWait(), delta)

	now = doRequests(throttler, 100, now, time.Second)
	assert.InDelta(t, time.Duration(0), throttler.calculateWait(), delta)
}

func TestThrottler_Bursts(t *testing.T) {
	throttler := NewThrottler(10_000, time.Hour, 2*time.Hour, logger)

	now := doRequests(throttler, 200, time.Now(), time.Millisecond)
	assert.InDelta(t, time.Minute+12*time.Second, throttler.calculateWait(), delta)

	now = doRequests(throttler, 100, now, time.Second)
	assert.InDelta(t, time.Minute+47*time.Second, throttler.calculateWait(), delta)
}

// Do numRequests many requests from start within interval
func doRequests(throttler *Throttler, numRequests int, start time.Time, interval time.Duration) time.Time {
	period := interval / time.Duration(numRequests-1)
	for i := 0; i < numRequests; i++ {
		throttler.add(start.Add(period * time.Duration(i)))
	}
	return start.Add(interval)
}
