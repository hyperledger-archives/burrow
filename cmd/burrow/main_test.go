package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBurrow(t *testing.T) {
	var outputCount int
	out := &output{
		PrintfFunc: func(format string, args ...interface{}) {
			outputCount++
		},
		LogfFunc: func(format string, args ...interface{}) {
			outputCount++
		},
		FatalfFunc: func(format string, args ...interface{}) {
			t.Fatalf("fatalf called by burrow cmd: %s", fmt.Sprintf(format, args...))
		},
	}
	app := burrow(out)
	// Basic smoke test for cli config
	assert.NoError(t, app.Run([]string{"burrow", "--version"}))
	assert.NoError(t, app.Run([]string{"burrow", "spec", "--name-prefix", "foo", "-f1"}))
	assert.NoError(t, app.Run([]string{"burrow", "configure"}))
	assert.NoError(t, app.Run([]string{"burrow", "start", "-h"}))
	assert.True(t, outputCount > 0)
}
