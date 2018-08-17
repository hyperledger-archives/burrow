package loggers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_sortKeyvals(t *testing.T) {
	keyvals := []interface{}{"foo", 3, "bar", 5}
	indices := map[string]int{"foo": 1, "bar": 0}
	sortKeyvals(indices, keyvals)
	assert.Equal(t, []interface{}{"bar", 5, "foo", 3}, keyvals)
}

func TestSortLogger(t *testing.T) {
	testLogger := newTestLogger()
	sortLogger := SortLogger(testLogger, "foo", "bar", "baz")
	sortLogger.Log([][]int{}, "bar", "foo", 3, "baz", "horse", "crabs", "cycle", "bar", 4, "ALL ALONE")
	sortLogger.Log("foo", 0)
	sortLogger.Log("bar", "foo", "foo", "baz")
	lines, err := testLogger.logLines(3)
	require.NoError(t, err)
	// non string keys sort after string keys, specified keys sort before unspecifed keys, specified key sort in order
	assert.Equal(t, [][]interface{}{
		{"foo", 3, "bar", 4, "baz", "horse", [][]int{}, "bar", "crabs", "cycle", "ALL ALONE"},
		{"foo", 0},
		{"foo", "baz", "bar", "foo"},
	}, lines)
}
