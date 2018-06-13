package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRange(t *testing.T) {
	start, end, err := parseRange("")
	require.NoError(t, err)
	assert.Equal(t, int64(0), start)
	assert.Equal(t, int64(-1), end)

	start, end, err = parseRange(":")
	require.NoError(t, err)
	assert.Equal(t, int64(0), start)
	assert.Equal(t, int64(-1), end)

	start, end, err = parseRange("0:")
	require.NoError(t, err)
	assert.Equal(t, int64(0), start)
	assert.Equal(t, int64(-1), end)

	start, end, err = parseRange(":-1")
	require.NoError(t, err)
	assert.Equal(t, int64(0), start)
	assert.Equal(t, int64(-1), end)

	start, end, err = parseRange("0:-1")
	require.NoError(t, err)
	assert.Equal(t, int64(0), start)
	assert.Equal(t, int64(-1), end)

	start, end, err = parseRange("123123:-123")
	require.NoError(t, err)
	assert.Equal(t, int64(123123), start)
	assert.Equal(t, int64(-123), end)
}
