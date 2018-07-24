package query

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatches(t *testing.T) {
	var (
		txDate = "2017-01-01"
		txTime = "2018-05-03T14:45:00Z"
	)

	testCases := []struct {
		s       string
		tags    map[string]interface{}
		err     bool
		matches bool
	}{
		{"tm.events.type='NewBlock'", map[string]interface{}{"tm.events.type": "NewBlock"}, false, true},

		{"tx.gas > 7", map[string]interface{}{"tx.gas": "8"}, false, true},
		{"tx.gas > 7 AND tx.gas < 9", map[string]interface{}{"tx.gas": "8"}, false, true},
		{"body.weight >= 3.5", map[string]interface{}{"body.weight": "3.5"}, false, true},
		{"account.balance < 1000.0", map[string]interface{}{"account.balance": "900"}, false, true},
		{"apples.kg <= 4", map[string]interface{}{"apples.kg": "4.0"}, false, true},
		{"body.weight >= 4.5", map[string]interface{}{"body.weight": fmt.Sprintf("%v", float32(4.5))}, false, true},
		{"oranges.kg < 4 AND watermellons.kg > 10", map[string]interface{}{"oranges.kg": "3", "watermellons.kg": "12"}, false, true},
		{"peaches.kg < 4", map[string]interface{}{"peaches.kg": "5"}, false, false},

		{"tx.date > DATE 2017-01-01", map[string]interface{}{"tx.date": time.Now().Format(DateLayout)}, false, true},
		{"tx.date = DATE 2017-01-01", map[string]interface{}{"tx.date": txDate}, false, true},
		{"tx.date = DATE 2018-01-01", map[string]interface{}{"tx.date": txDate}, false, false},

		{"tx.time >= TIME 2013-05-03T14:45:00Z", map[string]interface{}{"tx.time": time.Now().Format(TimeLayout)}, false, true},
		{"tx.time = TIME 2013-05-03T14:45:00Z", map[string]interface{}{"tx.time": txTime}, false, false},

		{"abci.owner.name CONTAINS 'Igor'", map[string]interface{}{"abci.owner.name": "Igor,Ivan"}, false, true},
		{"abci.owner.name CONTAINS 'Igor'", map[string]interface{}{"abci.owner.name": "Pavel,Ivan"}, false, false},
	}

	for _, tc := range testCases {
		q, err := New(tc.s)
		if !tc.err {
			require.Nil(t, err)
		}

		if tc.matches {
			assert.True(t, q.Matches(TagMap(tc.tags)), "Query '%s' should match %v", tc.s, tc.tags)
		} else {
			assert.False(t, q.Matches(TagMap(tc.tags)), "Query '%s' should not match %v", tc.s, tc.tags)
		}
	}
}

func TestMustParse(t *testing.T) {
	assert.Panics(t, func() { MustParse("=") })
	assert.NotPanics(t, func() { MustParse("tm.events.type='NewBlock'") })
}

func TestConditions(t *testing.T) {
	txTime, err := time.Parse(time.RFC3339, "2013-05-03T14:45:00Z")
	require.NoError(t, err)

	testCases := []struct {
		s          string
		conditions []Condition
	}{
		{s: "tm.events.type='NewBlock'", conditions: []Condition{{Tag: "tm.events.type", Op: OpEqual, Operand: "NewBlock"}}},
		{s: "tx.gas > 7 AND tx.gas < 9", conditions: []Condition{{Tag: "tx.gas", Op: OpGreater, Operand: int64(7)}, {Tag: "tx.gas", Op: OpLess, Operand: int64(9)}}},
		{s: "tx.time >= TIME 2013-05-03T14:45:00Z", conditions: []Condition{{Tag: "tx.time", Op: OpGreaterEqual, Operand: txTime}}},
	}

	for _, tc := range testCases {
		q, err := New(tc.s)
		require.Nil(t, err)

		assert.Equal(t, tc.conditions, q.Conditions())
	}
}
