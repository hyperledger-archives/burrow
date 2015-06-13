package test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	vegeta "github.com/tsenart/vegeta/lib"
	"testing"
	"time"
)

const (
	DURATION = 3
	PER_SEC  = 200
)

// Coarse flood testing just to ensure that http server
// does not crash.
func TestHttpFlooding(t *testing.T) {
	serveProcess := NewServeScumbag()
	errSS := serveProcess.Start()
	assert.NoError(t, errSS, "Scumbag-ed!")
	t.Logf("Flooding http requests.")
	err := runHttp()
	if err == nil {
		t.Logf("HTTP test: A total of %d http GET messages sent succesfully over %d seconds.\n", DURATION*PER_SEC, DURATION)
	}
	stopC := serveProcess.StopEventChannel()
	errStop := serveProcess.Stop(0)
	<-stopC
	assert.NoError(t, errStop, "Scumbag-ed!")

}

func runHttp() error {
	rate := uint64(PER_SEC) // per second
	duration := DURATION * time.Second
	targeter := vegeta.NewStaticTargeter(&vegeta.Target{
		Method: "GET",
		URL:    "http://localhost:1337/scumbag",
	})
	attacker := vegeta.NewAttacker()
	var results vegeta.Results
	for res := range attacker.Attack(targeter, rate, duration) {
		results = append(results, res)
	}
	metrics := vegeta.NewMetrics(results)
	if len(metrics.Errors) > 0 {
		return fmt.Errorf("Errors: %v\n", metrics.Errors)
	}
	return nil
}
