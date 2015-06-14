package erisdb

import (
	ep "github.com/eris-ltd/erisdb/erisdb/pipe"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Test empty query.
func TestEmptyQuery(t *testing.T) {
	arr, err := _parseQuery("")
	assert.NoError(t, err)
	assert.Nil(t, arr, "Array should be nil")
}

// Test no colon separated filter.
func TestQueryNoColonSeparator(t *testing.T) {
	_, err := _parseQuery("test")
	assert.Error(t, err, "Should detect missing colon.")
}

// Test no colon separated filter and proper filters mixed.
func TestQueryNoColonSeparatorMulti(t *testing.T) {
	_, err := _parseQuery("test test1:24 test2")
	assert.Error(t, err, "Should detect missing colon.")
}

// Test how it handles a query with op and value empty.
func TestQueryOpEmptyValueEmpty(t *testing.T) {
	assertFilter(t, "test:", "test", "==", "")
}

// Test how it handles a query with an empty op but a proper value.
func TestQueryOpEmptyValue(t *testing.T) {
	assertFilter(t, "test:val", "test", "==", "val")
}

// Test the '>' operator.
func TestQueryGT(t *testing.T) {
	testOp(">", t)
}

// Test the '<' operator.
func TestQueryLT(t *testing.T) {
	testOp("<", t)
}

// Test the '>=' operator.
func TestQueryGTEQ(t *testing.T) {
	testOp(">=", t)
}

// Test the '<=' operator.
func TestQueryLTEQ(t *testing.T) {
	testOp("<=", t)
}

// Test the '==' operator.
func TestQueryEQ(t *testing.T) {
	testOp("==", t)
}

// Test the '!=' operator.
func TestQueryNEQ(t *testing.T) {
	testOp("!=", t)
}

func TestCombined(t *testing.T) {
	q := "balance:>=5 sequence:<8"
	arr, err := _parseQuery(q)
	assert.NoError(t, err)
	assert.Len(t, arr, 2)
	f0 := arr[0]
	assert.Equal(t, f0.Field, "balance")
	assert.Equal(t, f0.Op, ">=")
	assert.Equal(t, f0.Value, "5")
	f1 := arr[1]
	assert.Equal(t, f1.Field, "sequence")
	assert.Equal(t, f1.Op, "<")
	assert.Equal(t, f1.Value, "8")

}

// Test a working range query.
func TestRangeQuery(t *testing.T) {
	assertRangeFilter(t, "5", "50", "5", "50")
}

// Test a working range-query with wildcard for lower bound.
func TestRangeQueryWildcardLB(t *testing.T) {
	assertRangeFilter(t, "*", "50", "min", "50")
}

// Test a working range-query with wildcard for upper bound.
func TestRangeQueryWildcardUB(t *testing.T) {
	assertRangeFilter(t, "5", "*", "5", "max")
}

// Test a working range-query with wildcard for upper and lower bound.
func TestRangeQueryWildcardULB(t *testing.T) {
	assertRangeFilter(t, "*", "*", "min", "max")
}

// Test a range query with no upper bounds term.
func TestRangeQueryBotchedMax(t *testing.T) {
	_, err := _parseQuery("test:5..")
	assert.Error(t, err, "Malformed range-query passed")
}

// Test a range query with no lower bounds term.
func TestRangeQueryBotchedMin(t *testing.T) {
	_, err := _parseQuery("test:..5")
	assert.Error(t, err, "Malformed range-query passed")
}

// Helpers.

func testOp(op string, t *testing.T) {
	assertFilter(t, "test:"+op+"33", "test", op, "33")
}

func assertFilter(t *testing.T, filter, field, op, val string) {
	arr, err := _parseQuery(filter)
	assert.NoError(t, err)
	assert.NotNil(t, arr)
	assert.Len(t, arr, 1)
	assert.Equal(t, arr[0], &ep.FilterData{field, op, val})
}

func assertRangeFilter(t *testing.T, min, max, res0, res1 string) {
	arr, err := _parseQuery("test:" + min + ".." + max)
	assert.NoError(t, err)
	assert.NotNil(t, arr)
	assert.Len(t, arr, 2)
	assert.Equal(t, arr[0], &ep.FilterData{"test", ">=", res0})
	assert.Equal(t, arr[1], &ep.FilterData{"test", "<=", res1})
}
