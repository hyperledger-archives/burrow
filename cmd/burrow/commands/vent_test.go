package commands

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRequestRate(t *testing.T) {
	requests, base, err := parseRequestRate("2334223/24h")
	require.NoError(t, err)
	assert.Equal(t, 2334223, requests)
	assert.Equal(t, time.Hour*24, base)

	requests, base, err = parseRequestRate("99990/24h")
	require.NoError(t, err)
	assert.Equal(t, 99_990, requests)
	assert.Equal(t, time.Hour*24, base)
}
