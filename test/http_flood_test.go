package test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	vegeta "github.com/tsenart/vegeta/lib"
	"testing"
	"time"
)

const (
	DURATION = 1
	PER_SEC  = 10
)

// Coarse flood testing just to ensure that http server
// does not crash.
func TestHttpFlooding(t *testing.T) {
	serveProcess := NewServeScumbag()
	errSS := serveProcess.Start()
	assert.NoError(t, errSS, "Scumbag-ed!")
	err := runHttp()
	if err == nil {
		fmt.Printf("HTTP test: A total of %d http GET messages sent succesfully over %d seconds.\n", DURATION*PER_SEC, DURATION)
	}
	errStop := serveProcess.Stop(time.Millisecond*1000)
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
	if(len(metrics.Errors) > 0){
		return fmt.Errorf("Errors: %v\n", metrics.Errors)
	}
	return nil
}
